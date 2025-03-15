package types

const (
	NUMFLOORS      = 4
	NUMBUTTONTYPE  = 3
	NUMHALLBUTTONS = 2
)

type ElevatorBehaviour int

const (
	EB_Idle ElevatorBehaviour = iota
	EB_DoorOpen
	EB_Moving
)

type ClearRequestVariant int

const (
	CV_All ClearRequestVariant = iota
	CV_InDirn
)

type MotorDirection int

const (
	MD_Up   MotorDirection = 1
	MD_Down MotorDirection = -1
	MD_Stop MotorDirection = 0
)

type Direction int

const (
	D_Up   Direction = 1
	D_Down Direction = -1
	D_Stop Direction = 0
)

type ButtonType int

const (
	BT_HallUp   ButtonType = 0
	BT_HallDown ButtonType = 1
	BT_Cab      ButtonType = 2
)

type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

type OrderEvent struct {
	ElevatorID string
	Completed  bool
	Orders     []ButtonEvent
}

type DirnBehaviourPair struct {
	Dirn      MotorDirection
	Behaviour ElevatorBehaviour
}

type ClearRequestCallback func(button ButtonType, floor int)

type OrderMatrix [NUMFLOORS][NUMBUTTONTYPE]bool

type GlobalOrderMap map[string]OrderMatrix

type Elevator struct {
	ID        string
	Floor     int
	Dirn      MotorDirection
	Requests  OrderMatrix
	Behaviour ElevatorBehaviour
	Avaliable bool
	Config    struct {
		ClearRequestVariant ClearRequestVariant
		DoorOpenDuration    float64
		TimeBetweenFloors   float64
	}
}

type Receipient int

const (
	All Receipient = iota
	Master
)

func (behaviour ElevatorBehaviour) ToString() string{
	behavList := []string{"idle", "doorOpen", "moving"}
	return behavList[int(behaviour)]
}

func (dirn MotorDirection) ToString() string{
	dirnList := []string{"down","stop","up"}
	return dirnList[dirn+1]
}

type NetworkMessage struct {
	MsgType    string
	MsgData    interface{}
	Receipient Receipient
}