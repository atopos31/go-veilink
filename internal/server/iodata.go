package server

import "sync"

// 保存listener输入输出的数据总大小
type IOdata struct {
	input  int64
	output int64
	inmux  sync.RWMutex
	outmux sync.RWMutex
}

func (io *IOdata) AddInput(n int64) {
	io.inmux.Lock()
	defer io.inmux.Unlock()
	io.input += n
}

func (io *IOdata) AddOutput(n int64) {
	io.outmux.Lock()
	defer io.outmux.Unlock()
	io.output += n
}

func (io *IOdata) GetInput() int64 {
	io.inmux.RLock()
	defer io.inmux.RUnlock()
	return io.input
}

func (io *IOdata) GetOutput() int64 {
	io.outmux.RLock()
	defer io.outmux.RUnlock()
	return io.output
}
