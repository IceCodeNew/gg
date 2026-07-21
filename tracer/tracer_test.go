package tracer

import (
	"os"
	"runtime"
	"syscall"
	"testing"
)

func TestWaitForInitialTraceStop(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	proc, err := os.StartProcess("/bin/true", []string{"true"}, &os.ProcAttr{
		Sys: &syscall.SysProcAttr{Ptrace: true},
	})
	if err != nil {
		t.Fatalf("start tracee: %v", err)
	}
	if err := waitForInitialTraceStop(proc.Pid); err != nil {
		_ = proc.Kill()
		t.Fatalf("wait for tracee: %v", err)
	}
	if err := syscall.PtraceDetach(proc.Pid); err != nil {
		_ = proc.Kill()
		t.Fatalf("detach tracee: %v", err)
	}
	state, err := proc.Wait()
	if err != nil {
		t.Fatalf("wait for detached process: %v", err)
	}
	if !state.Success() {
		t.Fatalf("detached process exited unsuccessfully: %v", state)
	}
}
