// Copyright 2021 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package native

import (
	"encoding/json"
	"math/big"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth/tracers"
)

//go:generate go run github.com/fjl/gencodec -type callFrame -field-override callFrameMarshaling -out gen_callframe_json.go

func init() {
	tracers.DefaultDirectory.Register("call2Tracer", newCallTracer2, false)
}

type call2Tracer struct {
	noopTracer
	addressesTouched []common.Address
	interrupt        atomic.Bool // Atomic flag to signal execution interruption
	reason           error       // Textual reason for the interruption
}

// newCall2Tracer returns a native go tracer which tracks
// call frames of a tx, and implements vm.EVMLogger.
func newCallTracer2(ctx *tracers.Context, _ json.RawMessage) (tracers.Tracer, error) {
	t, err := newCall2TracerObject(ctx)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func newCall2TracerObject(ctx *tracers.Context) (*call2Tracer, error) {
	// First callframe contains tx context info
	// and is populated on start and end.
	return &call2Tracer{addressesTouched: make([]common.Address, 0)}, nil
}

// OnEnter is called when EVM enters a new scope (via call, create or selfdestruct).
func (t *call2Tracer) CaptureEnter(typ vm.OpCode, _ common.Address, to common.Address, _ []byte, _ uint64, _ *big.Int) {
	if !(typ == vm.DELEGATECALL || typ == vm.CALL) {
		return
	}
	t.addressesTouched = append(t.addressesTouched, to)
	// t.depth = depth

	// // Skip if tracing was interrupted
	// if t.interrupt.Load() {
	// 	return
	// }
	// if !(vm.OpCode(typ) == vm.DELEGATECALL || vm.OpCode(typ) == vm.CALL) {
	// 	return
	// }
	// toCopy := to
	// call := callFrame2{
	// 	Type:  vm.OpCode(typ),
	// 	To:    &toCopy,
	// }
	// if depth == 0 {
	// 	call.Gas = t.gasLimit
	// }
	// t.callstack = append(t.callstack, call)
}

// GetResult returns the json-encoded nested list of call traces, and any
// error arising from the encoding or forceful termination (via `Stop`).
func (t *call2Tracer) GetResult() (json.RawMessage, error) {
	res, err := json.Marshal(t.addressesTouched)
	if err != nil {
		return nil, err
	}
	return res, t.reason
}

// Stop terminates execution of the tracer at the first opportune moment.
func (t *call2Tracer) Stop(err error) {
	t.reason = err
	t.interrupt.Store(true)
}
