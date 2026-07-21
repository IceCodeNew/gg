package tracer

import (
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"testing"
)

func TestWaitForInitialTraceStop(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	truePath, err := exec.LookPath("true")
	if err != nil {
		t.Fatalf("find true executable: %v", err)
	}
	proc, err := os.StartProcess(truePath, []string{"true"}, &os.ProcAttr{
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

func TestWaitForInitialTraceStopRetriesEINTR(t *testing.T) {
	const pid = 42
	calls := 0
	err := waitForInitialTraceStopWith(pid, func(gotPID int, status *syscall.WaitStatus, options int, _ *syscall.Rusage) (int, error) {
		calls++
		if gotPID != pid {
			t.Fatalf("wait called with pid %d, want %d", gotPID, pid)
		}
		if options != syscall.WALL|syscall.WUNTRACED {
			t.Fatalf("wait called with options %#x", options)
		}
		if calls == 1 {
			return 0, syscall.EINTR
		}
		*status = syscall.WaitStatus(uint32(syscall.SIGTRAP)<<8 | 0x7f)
		return pid, nil
	})
	if err != nil {
		t.Fatalf("wait for tracee: %v", err)
	}
	if calls != 2 {
		t.Fatalf("wait called %d times, want 2", calls)
	}
}
