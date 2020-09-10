package m3api

import "github.com/freddy33/qsm-go/model/m3point"

const (
	HttpEnvIdKey = "QsmEnvId"
)

func PointMsgToPoint(pMsg *PointMsg) m3point.Point {
	return m3point.Point{m3point.CInt(pMsg.X), m3point.CInt(pMsg.Y), m3point.CInt(pMsg.Z)}
}

func PointToPointMsg(p m3point.Point) *PointMsg {
	return &PointMsg{X: int32(p.X()), Y: int32(p.Y()), Z: int32(p.Z())}
}
