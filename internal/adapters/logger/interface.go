package logger

// LoggerAdapter не получится объявлять по месту использования как я делал ранее из-за метода With
// с другой стороны это как будто и не очень удобно было бы делать, так как логгер используется много где и каждый раз писать мини-интерфейс не очень хочется
type LoggerAdapter interface {
	Info(string, ...any)
	Error(string, ...any)
	Warn(string, ...any)
	Debug(string, ...any)
	With(...any) LoggerAdapter
}
