package eventdefinition

type EventNotification struct {
	Level   string
	Message string
}

const (
	MINOR  = "MINOR"
	MEDIUM = "MEDIUM"
	MAJOR  = "MAJOR"
)

func MapLevelToIcon(level string) string {
	switch level {
	case MINOR:
		return "🐞"
	case MEDIUM:
		return "️⚠️"
	case MAJOR:
		return "❗️"
	}
	return "❔"
}
