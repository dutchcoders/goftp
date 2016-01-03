package goftp

// FTP status codes https://en.wikipedia.org/wiki/List_of_FTP_server_return_codes
const (
	CodeRestartMarker             = 110
	CodeServiceReadyNNNminutes    = 120
	CodeDataConnectionAlreadyOpen = 125
	CodeFileStatusOk              = 150

	CodeCommandOk                              = 200
	CodeCommandNotImplementedInfo              = 202
	CodeSystemStatus                           = 211
	CodeDirectoryStatus                        = 212
	CodeFileStatus                             = 213
	CodeHelpMessage                            = 214
	CodeSystemType                             = 215
	CodeServiceReadyForNewUser                 = 220
	CodeServiceClosingControlConnection        = 221
	CodeDataConnectionOpenNoTransferInProgress = 225
	CodeClosingDataConnection                  = 226
	CodeEnteringPassiveMode                    = 227
	CodeEnteringLongPassiveMode                = 228
	CodeEnteringExtendedPassiveMode            = 229
	CodeUserLoggedIn                           = 230
	CodeUserLoggedOut                          = 231
	CodeLogoutNoted                            = 232
	CodeAuthMechanismAccepted                  = 234
	CodeRequestedFileActionOk                  = 250
	CodePathnameCreated                        = 257

	CodeUserNameOkNeedPassword     = 331
	CodeNeedAccountForLogin        = 332
	CodeRequestedFileActionPending = 350

	CodeServiceNotAvaliable         = 421
	CodeCantOpenDataConnection      = 425
	CodeConnectionClosed            = 426
	CodeInvalidUsernameOrPassword   = 430
	CodeRequstedHostUnavaliable     = 434
	CodeRequestedFileActionNotTaken = 450
	CodeRequestedActionAborted      = 451
	CodeRequestedActionNotTaken     = 452

	CodeSyntaxError                        = 501
	CodeCommandNotImplementedError         = 502
	CodeBadSequenceOfCommandsError         = 503
	CodeCommandNotImplementedForParamError = 504
	CodeNotLoggedInError                   = 530
	CodeNeedAccountForStoringError         = 532
	CodeFileUnavailableError               = 550
	CodePageTypeunknownError               = 551
	CodeExceededStorageAllocationError     = 552
	CodeFileNameNotAllowedError            = 553

	CodeIntegrityProtectedReply                   = 631
	CodeConfidentialityAndIntegrityProtectedReply = 632
	CodeConfidentialityProtectedReply             = 633
)
