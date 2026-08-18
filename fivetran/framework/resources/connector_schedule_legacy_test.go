package resources

import (
	"testing"

	"github.com/fivetran/go-fivetran/connections"
)

func TestLegacyCompatibleScheduleFields(t *testing.T) {
	tests := []struct {
		name              string
		schedule          *connections.ConnectorSchedule
		wantSyncFrequency *int
		wantDailySyncTime *string
	}{
		{
			name: "interval with omitted days maps to sync frequency",
			schedule: &connections.ConnectorSchedule{
				ScheduleType: strPtr("INTERVAL"),
				Interval:     intPtr(60),
			},
			wantSyncFrequency: intPtr(60),
		},
		{
			name: "interval with all days maps to sync frequency",
			schedule: &connections.ConnectorSchedule{
				ScheduleType: strPtr("INTERVAL"),
				Interval:     intPtr(120),
				DaysOfWeek: []string{
					"MONDAY",
					"TUESDAY",
					"WEDNESDAY",
					"THURSDAY",
					"FRIDAY",
					"SATURDAY",
					"SUNDAY",
				},
			},
			wantSyncFrequency: intPtr(120),
		},
		{
			name: "daily time of day maps to daily sync time",
			schedule: &connections.ConnectorSchedule{
				ScheduleType: strPtr("TIME_OF_DAY"),
				TimeOfDay:    strPtr("09:00"),
			},
			wantSyncFrequency: intPtr(1440),
			wantDailySyncTime: strPtr("09:00"),
		},
		{
			name: "daily time of day with seconds maps to daily sync time",
			schedule: &connections.ConnectorSchedule{
				ScheduleType: strPtr("TIME_OF_DAY"),
				TimeOfDay:    strPtr("09:00:00"),
			},
			wantSyncFrequency: intPtr(1440),
			wantDailySyncTime: strPtr("09:00"),
		},
		{
			name: "unsupported interval does not map",
			schedule: &connections.ConnectorSchedule{
				ScheduleType: strPtr("INTERVAL"),
				Interval:     intPtr(45),
			},
		},
		{
			name: "non-hourly time of day does not map",
			schedule: &connections.ConnectorSchedule{
				ScheduleType: strPtr("TIME_OF_DAY"),
				TimeOfDay:    strPtr("09:30"),
			},
		},
		{
			name: "interval with subset days does not map",
			schedule: &connections.ConnectorSchedule{
				ScheduleType: strPtr("INTERVAL"),
				Interval:     intPtr(60),
				DaysOfWeek:   []string{"MONDAY", "WEDNESDAY", "FRIDAY"},
			},
		},
		{
			name: "time of day with subset days does not map",
			schedule: &connections.ConnectorSchedule{
				ScheduleType: strPtr("TIME_OF_DAY"),
				TimeOfDay:    strPtr("09:00"),
				DaysOfWeek:   []string{"MONDAY", "WEDNESDAY", "FRIDAY"},
			},
		},
		{
			name: "cron does not map",
			schedule: &connections.ConnectorSchedule{
				ScheduleType: strPtr("CRON"),
				Cron:         strPtr("0 9 * * 1-5"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSyncFrequency, gotDailySyncTime := legacyCompatibleScheduleFields(tt.schedule)

			assertIntPointer(t, gotSyncFrequency, tt.wantSyncFrequency)
			assertStringPointer(t, gotDailySyncTime, tt.wantDailySyncTime)
		})
	}
}

func strPtr(value string) *string {
	return &value
}

func intPtr(value int) *int {
	return &value
}

func assertIntPointer(t *testing.T, got, want *int) {
	t.Helper()

	if got == nil || want == nil {
		if got != want {
			t.Fatalf("got %v, want %v", got, want)
		}
		return
	}

	if *got != *want {
		t.Fatalf("got %v, want %v", *got, *want)
	}
}

func assertStringPointer(t *testing.T, got, want *string) {
	t.Helper()

	if got == nil || want == nil {
		if got != want {
			t.Fatalf("got %v, want %v", got, want)
		}
		return
	}

	if *got != *want {
		t.Fatalf("got %v, want %v", *got, *want)
	}
}
