package store

import (
	"context"
	"testing"
)

func TestDeleteActionEventsByPrefixClearsTodayOnly(t *testing.T) {
	st, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	ctx := context.Background()

	// 今天的幂等键（应被清除）与明天的（应保留）
	if _, err := st.RecordActionEvent(ctx, "schedule:1:20260818:stop", 1, "schedule_stop", "executed", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := st.RecordActionEvent(ctx, "schedule:1:20260818:start", 1, "schedule_start", "executed", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := st.RecordActionEvent(ctx, "schedule:1:20260819:stop", 1, "schedule_stop", "executed", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := st.RecordActionEvent(ctx, "schedule:2:20260818:stop", 2, "schedule_stop", "executed", ""); err != nil {
		t.Fatal(err)
	}

	// 清除账号 1 今天的调度幂等键
	if err := st.DeleteActionEventsByPrefix(ctx, "schedule:1:20260818:"); err != nil {
		t.Fatal(err)
	}

	// 账号1今天已清空，新执行应重新记录（fresh=true）
	fresh, err := st.RecordActionEvent(ctx, "schedule:1:20260818:stop", 1, "schedule_stop", "attempting", "")
	if err != nil {
		t.Fatal(err)
	}
	if !fresh {
		t.Fatal("期望今天已清除后可重新记录, 实际 fresh=false")
	}

	// 其余键不受影响
	freshStart, err := st.RecordActionEvent(ctx, "schedule:1:20260818:start", 1, "schedule_start", "attempting", "")
	if err != nil {
		t.Fatal(err)
	}
	if !freshStart {
		t.Fatal("今天的 start 键也应被清除, 实际 fresh=false")
	}
	// 明天的键保留
	freshTomorrow, err := st.RecordActionEvent(ctx, "schedule:1:20260819:stop", 1, "schedule_stop", "attempting", "")
	if err != nil {
		t.Fatal(err)
	}
	if freshTomorrow {
		t.Fatal("明天的键应保留, 实际 fresh=true (说明被误删)")
	}
	// 其他账号的键保留
	freshOther, err := st.RecordActionEvent(ctx, "schedule:2:20260818:stop", 2, "schedule_stop", "attempting", "")
	if err != nil {
		t.Fatal(err)
	}
	if freshOther {
		t.Fatal("其他账号的键应保留, 实际 fresh=true (说明被误删)")
	}
}