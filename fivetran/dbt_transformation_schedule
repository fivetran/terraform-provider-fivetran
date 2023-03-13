package fivetran

// DbtTransformationSchedule builds dbt transformation management, dbt transformation schedule
// Ref. TODO url
type DbtTransformationSchedule struct {
	scheduleType *string
	daysOfWeek   []string
	timeOfDay    *time.Time
}


type dbtTransformationScheduleRequest struct {
	ScheduleType    *string
	DaysOfWeek      []string
	TimeOfDay       *time.Time
}

type dbtTransformationScheduleResponse struct {
	ScheduleType    *string
	DaysOfWeek      []string
	Interval        *string // maybe add this to request model also
	TimeOfDay       *time.Time
}

func NewDbtTransformationSchedule() *DbtTransformationSchedule {
	return &DbtTransformationSchedule{}
}

func (cc *DbtTransformationSchedule) request() *dbtTransformationScheduleRequest {
	
	return *dbtTransformationScheduleRequest {
		ScheduleType:       cc.scheduleType
		DaysOfWeek:         cc.DaysOfWeek
		TimeOfDay:          cc.timeOfDay
	}
}

func (cc *DbtTransformationSchedule) ScheduleType(value string) *DbtTransformationSchedule {
	cc.scheduleType = &value
	return cc
}

func (cc *DbtTransformationSchedule) DaysOfWeek(value []string) *DbtTransformationSchedule {
	cc.daysOfWeek = value
	return cc
}

func (cc *DbtTransformationSchedule) TimeOfDay(value time.Time) *DbtTransformationSchedule {
	cc.timeOfDay = &value
	return cc
}


