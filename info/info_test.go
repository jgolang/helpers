package info

import "testing"

type test struct {
	Test string
	Info Info
}

func TestInfo_Set(t *testing.T) {
	info := &Info{}
	type args struct {
		key   string
		value interface{}
	}
	tests := []struct {
		name string
		data *Info
		args args
	}{
		{
			name: "ok",
			data: info,
			args: args{
				key:   "test",
				value: "value",
			},
		},
		{
			name: "ok",
			data: info,
			args: args{
				key:   "test2",
				value: "value2",
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.data.Set(tt.args.key, tt.args.value)
			t.Log(tt.data)
			t.Log(info.GetString(tt.args.key))
		})
	}
}

func TestStructAdditonalInfo(t *testing.T) {
	var info test
	info.Info.Set("key", "value")
	t.Log(info.Info)
	t.Log(info.Info.GetString("key"))
}
