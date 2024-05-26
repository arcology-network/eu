package execution

import (
	"encoding/json"
	"fmt"
)

type ExecutionLog struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (this *ExecutionLog) GetByKey() string {
	return this.Key
}

func (this *ExecutionLog) GetValue() string {
	return this.Value
}

type ExecutionLogs struct {
	Txhash [32]byte       `json:"txhash"`
	Logs   []ExecutionLog `json:"this"`
}

type ExecutionLogsMessage struct {
	Logs   ExecutionLogs
	Height uint64
	Round  uint64
	Msgid  uint64
}

func GetAssert(ret []byte) string {
	startIdx := 4 + 32 + 32
	pattern := []byte{8, 195, 121, 160}
	if ret != nil || len(ret) > startIdx {
		starts := ret[:4]
		if string(pattern) == string(starts) {
			remains := ret[startIdx:]
			end := 0
			for i := range remains {
				if remains[i] == 0 {
					end = i
					break
				}
			}
			return string(remains[:end])
		}
	}
	return ""
}

func NewExecutionLogs() *ExecutionLogs {
	return &ExecutionLogs{
		Logs: []ExecutionLog{},
	}
}

func (this *ExecutionLogs) Transform(key, value string) {
	this.Logs = append(this.Logs, ExecutionLog{
		Key:   key,
		Value: value,
	})
}
func (this *ExecutionLogs) Appends(log []ExecutionLog) {
	this.Logs = append(this.Logs, log...)
}

func (this *ExecutionLogs) Marshal() (string, error) {
	data, err := json.Marshal(this)
	return fmt.Sprintf("%v", string(data)), err
}

func (this *ExecutionLogs) UnMarshal(data string) error {

	return json.Unmarshal([]byte(data), this)
}
