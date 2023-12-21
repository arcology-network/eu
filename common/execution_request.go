package common

// type ExecutorRequest struct {
// 	Sequences   []*Sequence
// 	Timestamp   *big.Int
// 	Parallelism uint64
// 	Debug       bool
// }

// func (this *ExecutorRequest) GobEncode() ([]byte, error) {
// 	executingSequences := ExecutingSequences(this.Sequences)
// 	executingSequencesData, err := executingSequences.Encode()
// 	if err != nil {
// 		return []byte{}, err
// 	}

// 	precedingsBytes := make([][]byte, len(this.Precedings))
// 	for i := range this.Precedings {
// 		precedings := Ptr2Arr(this.Precedings[i])
// 		precedingsBytes[i] = ethCommon.Hashes(precedings).Encode()
// 	}

// 	timeStampData := []byte{}
// 	if this.Timestamp != nil {
// 		timeStampData = this.Timestamp.Bytes()
// 	}

// 	data := [][]byte{
// 		executingSequencesData,
// 		encoding.Byteset(precedingsBytes).Encode(),
// 		ethCommon.Hashes(this.PrecedingHash).Encode(),
// 		timeStampData,
// 		common.Uint64ToBytes(this.Parallelism),
// 		codec.Bool(this.Debug).Encode(),
// 	}
// 	return encoding.Byteset(data).Encode(), nil
// }

// func (this *ExecutorRequest) GobDecode(data []byte) error {
// 	fields := encoding.Byteset{}.Decode(data)
// 	msgResults, err := new(ExecutingSequences).Decode(fields[0])
// 	if err != nil {
// 		return err
// 	}
// 	this.Sequences = msgResults

// 	precedingsBytes := encoding.Byteset{}.Decode(fields[1])
// 	this.Precedings = make([][]*ethCommon.Hash, len(precedingsBytes))
// 	for i := range precedingsBytes {
// 		this.Precedings[i] = Arr2Ptr(ethCommon.Hashes([]ethCommon.Hash{}).Decode(precedingsBytes[i]))
// 	}

// 	this.PrecedingHash = ethCommon.Hashes([]ethCommon.Hash{}).Decode(fields[2])
// 	//if len(fields[3]) > 0 {
// 	this.Timestamp = new(big.Int).SetBytes(fields[3])
// 	//}
// 	//if len(fields[4]) > 0 {
// 	this.Parallelism = common.BytesToUint64(fields[4])
// 	//}
// 	this.Debug = bool(codec.Bool(this.Debug).Decode(fields[5]).(codec.Bool))
// 	return nil
// }
