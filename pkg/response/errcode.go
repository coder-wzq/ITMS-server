package response

const (
	CodeSuccess int = 0

	// Auth error codes (01xxx)
	CodeAuthParamErr     int = 1001 // parameter error (01001 in docs)
	CodeAuthTokenInvalid int = 1002 // invalid/expired token
	CodeAuthTokenMissing int = 1003 // missing authorization header
	CodeAuthForbidden    int = 1004 // forbidden
	CodeAuthNoPermission int = 1005 // insufficient permissions
	CodeAuthLoginFail    int = 1101 // username or password incorrect
	CodeAuthUserDisabled int = 1102 // user disabled
	CodeAuthUserLocked   int = 1103 // user locked
	CodeAuthServerErr    int = 1500 // internal server error

	// User error codes (02xxx)
	CodeUserNotFound     int = 2001 // user not found
	CodeUserDuplicate    int = 2101 // username already exists
	CodeUserBatchLimit   int = 2102 // batch operation exceeds limit
	CodeUserSelfDelete   int = 2103 // cannot delete self
	CodeUserSystemDelete int = 2104 // cannot delete system user

	// Director error codes (04xxx)
	CodeDirSessionNotFound int = 4001 // session not found
	CodeDirInvalidState    int = 4101 // invalid state transition
	CodeDirInvalidAction   int = 4102 // invalid action
	CodeDirRoomActive      int = 4103 // room has active session

	// Task error codes (06xxx)
	CodeTaskNotFound    int = 6001 // task not found
	CodeTaskDuplicate   int = 6101 // task number already exists
	CodeTaskImportFail  int = 6102 // import failed
	CodeTaskVersionFail int = 6103 // version conflict
	CodeTaskNotDraft    int = 6104 // task not in draft status

	// MPT error codes (07xxx)
	CodeMPTScenarioNotFound int = 7001 // scenario not found
	CodeMPTEntityLimit      int = 7101 // entity count exceeds limit

	// Sim proxy error codes (08xxx)
	CodeSimConnErr    int = 8101 // simulation connection error
	CodeSimTimeoutErr int = 8102 // simulation command timeout

	// Voice error codes (10xxx)
	CodeVoiceChannelDuplicate int = 10101 // channel already in use

	// Dict error codes (11xxx)
	CodeDictForceDuplicate  int = 11101 // force name already exists
	CodeDictUnitNotFound    int = 11102 // command unit not found
	CodeDictUnitHasChildren int = 11103 // unit has children cannot delete
	CodeDictEquipmentCheck  int = 11104 // equipment check constraint failed

	// Agent error codes (12xxx)
	CodeAgentGenFail int = 12101 // AI generation failed

	// Report error codes (13xxx)
	CodeReportParseFail int = 13101 // BIN file parse failed

	// Record error codes (14xxx)
	CodeRecordNotFound int = 14101 // record not found

	// Admin error codes (15xxx)
	CodeAdminNodeNotFound int = 15001 // node not found

	// Rate limit
	CodeTooManyRequests int = 1006 // rate limit exceeded
)

var codeMessage = map[int]string{
	CodeSuccess:              "success",
	CodeAuthParamErr:         "parameter error",
	CodeAuthTokenInvalid:     "invalid or expired token",
	CodeAuthTokenMissing:     "missing authorization header",
	CodeAuthForbidden:        "forbidden",
	CodeAuthNoPermission:     "insufficient permissions",
	CodeAuthLoginFail:        "username or password incorrect",
	CodeAuthUserDisabled:     "user has been disabled",
	CodeAuthUserLocked:       "user is locked due to too many login attempts",
	CodeAuthServerErr:        "internal server error",
	CodeTooManyRequests:      "too many requests, please try again later",
	CodeUserNotFound:         "user not found",
	CodeUserDuplicate:        "username already exists",
	CodeUserBatchLimit:       "batch operation exceeds limit (max 100)",
	CodeUserSelfDelete:       "cannot delete current user",
	CodeUserSystemDelete:     "cannot delete system administrator",
	CodeDirSessionNotFound:   "training session not found",
	CodeDirInvalidState:      "invalid state transition",
	CodeDirInvalidAction:     "invalid action",
	CodeDirRoomActive:        "room already has an active session",
	CodeTaskNotFound:         "task not found",
	CodeTaskDuplicate:        "task number already exists",
	CodeTaskImportFail:       "import failed",
	CodeTaskVersionFail:      "version conflict",
	CodeTaskNotDraft:         "task must be in draft status",
	CodeMPTScenarioNotFound:  "MPT scenario not found",
	CodeMPTEntityLimit:       "entity count exceeds per-side limit (max 200)",
	CodeSimConnErr:           "simulation engine connection failed",
	CodeSimTimeoutErr:        "simulation command timeout",
	CodeVoiceChannelDuplicate: "radio channel already in use",
	CodeDictForceDuplicate:   "force name already exists",
	CodeDictUnitNotFound:     "command unit not found",
	CodeDictUnitHasChildren:  "cannot delete unit with children",
	CodeDictEquipmentCheck:   "total >= available + maintenance constraint failed",
	CodeAgentGenFail:         "AI generation failed",
	CodeReportParseFail:      "BIN file parse failed",
	CodeRecordNotFound:       "training record not found",
	CodeAdminNodeNotFound:    "admin node not found",
}

func GetMessage(code int) string {
	if msg, ok := codeMessage[code]; ok {
		return msg
	}
	return "unknown error"
}
