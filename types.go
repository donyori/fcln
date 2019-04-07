package main

type Batch struct {
	Parent   string
	RegFiles []string
	Dirs     []string
}

type BatchList []*Batch

func (bl BatchList) Len() int {
	return len(bl)
}

func (bl BatchList) Less(i, j int) bool {
	return bl[i].Parent < bl[j].Parent
}

func (bl BatchList) Swap(i, j int) {
	bl[i], bl[j] = bl[j], bl[i]
}
