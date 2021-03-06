package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/donyori/goscut"
	"github.com/donyori/gotfp"
)

type Re struct {
	underlying *regexp.Regexp
}

type ConstraintInfoPattern struct {
	IsEmpty                  bool            `json:"is_empty,omitempty"`
	MinSize                  uint64          `json:"min_size,omitempty"`
	MaxSize                  uint64          `json:"max_size,omitempty"`
	EarliestModificationTime *time.Time      `json:"earliest_modification_time,omitempty"`
	LatestModificationTime   *time.Time      `json:"latest_modification_time,omitempty"`
	Permissions              [][]os.FileMode `json:"permissions,omitempty"`
}

type ConstraintFilePattern struct {
	ConstraintInfoPattern
	Ops []string `json:"ops,omitempty"`
}

type ConstraintFilePatternBatch struct {
	Default  []ConstraintFilePattern `json:"default,omitempty"`
	Dirs     []ConstraintFilePattern `json:"dirs,omitempty"`
	RegFiles []ConstraintFilePattern `json:"reg_files,omitempty"`
	Symlinks []ConstraintFilePattern `json:"symlinks,omitempty"`
}

type FilePattern struct {
	ConstraintInfoPattern
	Path         *Re                         `json:"path,omitempty"`
	Basename     *Re                         `json:"basename,omitempty"`
	CstrParent   *ConstraintFilePattern      `json:"cstr_parent,omitempty"`
	CstrSiblings *ConstraintFilePatternBatch `json:"cstr_siblings,omitempty"`
}

type PatternBatch struct {
	Default  []FilePattern `json:"default,omitempty"`
	Dirs     []FilePattern `json:"dirs,omitempty"`
	RegFiles []FilePattern `json:"reg_files,omitempty"`
	Symlinks []FilePattern `json:"symlinks,omitempty"`
}

var (
	skipPatternBatch   *PatternBatch
	removePatternBatch *PatternBatch

	loadPatternBatchesOnce sync.Once
)

func (re *Re) Match(s string) bool {
	if re == nil || re.underlying == nil {
		return false
	}
	return re.underlying.MatchString(s)
}

func (re *Re) String() string {
	if re == nil || re.underlying == nil {
		return "<nil>"
	}
	return re.underlying.String()
}

func (re *Re) MarshalText() ([]byte, error) {
	return []byte(re.String()), nil
}

func (re *Re) UnmarshalText(text []byte) error {
	r, err := regexp.Compile(string(text))
	if err != nil {
		return err
	}
	re.underlying = r
	return nil
}

func (cip *ConstraintInfoPattern) MatchInfo(info os.FileInfo) bool {
	if cip == nil {
		return true
	}
	if info == nil {
		return false
	}
	if cip.IsEmpty || cip.MinSize > 0 || cip.MaxSize > 0 {
		size := uint64(info.Size())
		if cip.IsEmpty && size != 0 {
			return false
		}
		if cip.MinSize > 0 && size < cip.MinSize {
			return false
		}
		if cip.MaxSize > 0 && size > cip.MaxSize {
			return false
		}
	}
	if cip.EarliestModificationTime != nil || cip.LatestModificationTime != nil {
		modTime := info.ModTime()
		if cip.EarliestModificationTime != nil &&
			cip.EarliestModificationTime.After(modTime) {
			return false
		}
		if cip.LatestModificationTime != nil &&
			cip.LatestModificationTime.Before(modTime) {
			return false
		}
	}
	if len(cip.Permissions) > 0 {
		perm := info.Mode().Perm()
		for _, disj := range cip.Permissions {
			if len(disj) == 0 {
				continue
			}
			ok := false
			for _, p := range disj {
				if p == perm {
					ok = true
					break
				}
			}
			if !ok {
				return false
			}
		}
	}
	return true
}

func (cfp *ConstraintFilePattern) Match(file, target *gotfp.FileInfo) bool {
	if cfp == nil {
		return true
	}
	if file == nil || !cfp.MatchInfo(file.Info) {
		return false
	}
	if len(cfp.Ops) > 0 {
		cutter := goscut.NewCutter()
		s := file.Info.Name()
		for _, op := range cfp.Ops {
			args, _, err := cutter.Cut(op)
			if err != nil {
				panic(err)
			}
			if len(args) == 0 {
				// Ignore empty op.
				continue
			}
			cmd := strings.ToLower(args[0])
			var arg string
			if len(args) >= 2 {
				arg = args[1]
			}
			switch cmd {
			case "set":
				s = arg
			case "setbasename", "set_basename":
				s = file.Info.Name()
			case "setpath", "set_path":
				s = file.Path
			case "addprefix", "add_prefix":
				s = arg + s
			case "addsuffix", "add_suffix":
				s = s + arg
			case "trim":
				s = strings.Trim(s, arg)
			case "trimprefix", "trim_prefix":
				s = strings.TrimPrefix(s, arg)
			case "trimsuffix", "trim_suffix":
				s = strings.TrimSuffix(s, arg)
			case "trimleft", "trim_left":
				s = strings.TrimLeft(s, arg)
			case "trimright", "trim_right":
				s = strings.TrimRight(s, arg)
			case "trimspace", "trim_space":
				s = strings.TrimSpace(s)
			case "eq", "equal", "equal_to":
				return s == arg
			case "eqt", "equal_to_target":
				if target == nil {
					return s == ""
				}
				return s == target.Path ||
					(target.Info != nil && s == target.Info.Name())
			default:
				panic(errors.New("fcln: unknown operation " + op))
			}
		}
		panic(errors.New("fcln: operations don't return a boolean at last"))
	}
	return true
}

func (fp *FilePattern) Match(file *gotfp.FileInfo,
	lctn *gotfp.LocationBatchInfo) bool {
	if fp == nil {
		return true
	}
	if file == nil {
		return false
	}
	if fp.Path != nil && !fp.Path.Match(file.Path) {
		return false
	}
	if fp.Basename != nil && !fp.Basename.Match(file.Info.Name()) {
		return false
	}
	if !fp.MatchInfo(file.Info) {
		return false
	}
	var batch *gotfp.Batch
	if lctn != nil {
		batch = lctn.Batch
	}
	if fp.CstrParent != nil &&
		(batch == nil || !fp.CstrParent.Match(&batch.Parent, file)) {
		return false
	}
	matchSiblings := func(cstr []ConstraintFilePattern,
		files [][]gotfp.FileInfo) bool {
		if batch == nil {
			return false
		}
		for i := range cstr {
			var ok bool
			for _, fs := range files {
				for j := range fs {
					fInfo := &fs[j]
					if fInfo == file {
						continue
					}
					if cstr[i].Match(fInfo, file) {
						ok = true
						break
					}
				}
				if ok {
					break
				}
			}
			if !ok {
				return false
			}
		}
		return true
	}
	if fp.CstrSiblings != nil {
		var defaultFiles [][]gotfp.FileInfo
		if len(fp.CstrSiblings.Default) > 0 {
			defaultFiles = make([][]gotfp.FileInfo, 0, 4)
		}
		if len(fp.CstrSiblings.Dirs) > 0 {
			if !matchSiblings(fp.CstrSiblings.Dirs,
				[][]gotfp.FileInfo{batch.Dirs}) {
				return false
			}
		} else if defaultFiles != nil {
			defaultFiles = append(defaultFiles, batch.Dirs)
		}
		if len(fp.CstrSiblings.RegFiles) > 0 {
			if !matchSiblings(fp.CstrSiblings.RegFiles,
				[][]gotfp.FileInfo{batch.RegFiles}) {
				return false
			}
		} else if defaultFiles != nil {
			defaultFiles = append(defaultFiles, batch.RegFiles)
		}
		if len(fp.CstrSiblings.Symlinks) > 0 {
			if !matchSiblings(fp.CstrSiblings.Symlinks,
				[][]gotfp.FileInfo{batch.Symlinks}) {
				return false
			}
		} else if defaultFiles != nil {
			defaultFiles = append(defaultFiles, batch.Symlinks)
		}
		if defaultFiles != nil {
			defaultFiles = append(defaultFiles, batch.Others)
			if !matchSiblings(fp.CstrSiblings.Default, defaultFiles) {
				return false
			}
		}
	}
	return true
}

func (pb *PatternBatch) Match(file *gotfp.FileInfo,
	lctn *gotfp.LocationBatchInfo) bool {
	if pb == nil || file == nil {
		return false
	}
	switch file.Cat {
	case gotfp.RegularFile:
		if len(pb.RegFiles) > 0 {
			return matchFile(pb.RegFiles, file, lctn)
		}
	case gotfp.Symlink:
		if len(pb.Symlinks) > 0 {
			return matchFile(pb.Symlinks, file, lctn)
		}
	case gotfp.Directory:
		if len(pb.Dirs) > 0 {
			return matchFile(pb.Dirs, file, lctn)
		}
	case gotfp.OtherFile:
		// Do nothing here.
	default:
		return false
	}
	return matchFile(pb.Default, file, lctn)
}

func matchFile(patterns []FilePattern,
	file *gotfp.FileInfo, lctn *gotfp.LocationBatchInfo) bool {
	for i := range patterns {
		if patterns[i].Match(file, lctn) {
			return true
		}
	}
	return false
}

func LazyLoadPatternBatches() {
	loadPatternBatchesOnce.Do(func() {
		patternBatchFilenames := []string{
			"skip.json",
			"remove.json",
		}
		patternBatches := []**PatternBatch{
			&skipPatternBatch,
			&removePatternBatch,
		}
		n := len(patternBatches)
		for i := 0; i < n; i++ {
			path := filepath.Join(patternsDir, patternBatchFilenames[i])
			data, err := ioutil.ReadFile(path)
			if err != nil {
				if os.IsNotExist(err) {
					*patternBatches[i] = nil
					continue
				}
				panic(err)
			}
			b := new(PatternBatch)
			err = json.Unmarshal(data, b)
			if err != nil {
				panic(err)
			}
			*patternBatches[i] = b
		}
	})
}
