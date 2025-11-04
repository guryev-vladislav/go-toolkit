package logger

const (
	MsgStart              = "start"
	MsgStartWithParams    = "start with params"
	MsgEnd                = "end"
	MsgPanicWasCatched    = "the panic was catched"
	MsgCompletesWithError = "%s completes with error"
	MsgSQLSelectWithError = "select from %s completes with error"
	MsgSQLInsertWithError = "insert into %s completes with error"
	MsgSQLUpdateWithError = "update %s completes with error"
	MsgSQLDeleteWithError = "delete from %s completes with error"
)
