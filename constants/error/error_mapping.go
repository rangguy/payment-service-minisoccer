package error

import (
	errorField "field-service/constants/error/field"
	errFieldSchedule "field-service/constants/error/fieldSchedule"
	errTime "field-service/constants/error/time"
)

func ErrMapping(err error) bool {
	allErrors := make([]error, 0)
	allErrors = append(append(append(GeneralErrors[:], errorField.FieldErrors[:]...), errFieldSchedule.FieldScheduleErrors[:]...), errTime.TimeErrors[:]...)

	for _, item := range allErrors {
		if err.Error() == item.Error() {
			return true
		}
	}

	return false
}
