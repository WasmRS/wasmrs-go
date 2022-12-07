package frames

type ErrCode uint32

const (
	ErrCodeReserved ErrCode = 0x00000000
	// The Setup frame is invalid for the server (it could be that the client is too recent for the old server). Stream ID MUST be 0.
	ErrCodeInvalidSetup ErrCode = 0x00000001
	// Some (or all) of the parameters specified by the client are unsupported by the server. Stream ID MUST be 0.
	ErrCodeUnsupportedSetup ErrCode = 0x00000002
	// The server rejected the setup, it can specify the reason in the payload. Stream ID MUST be 0.
	ErrCodeRejectedSetup ErrCode = 0x00000003
	// The server rejected the resume, it can specify the reason in the payload. Stream ID MUST be 0.
	ErrCodeRejectedResume ErrCode = 0x00000004
	// The connection is being terminated. Stream ID MUST be 0. Sender or Receiver of this frame MAY close the connection immediately without waiting for outstanding streams to terminate.
	ErrCodeConnectionError ErrCode = 0x00000101
	// The connection is being terminated. Stream ID MUST be 0. Sender or Receiver of this frame MUST wait for outstanding streams to terminate before closing the connection. New requests MAY not be accepted.
	ErrCodeConnectionClose ErrCode = 0x00000102
	// Application layer logic generating a Reactive Streams onError event. Stream ID MUST be > 0.
	ErrCodeApplicationError ErrCode = 0x00000201
	// Despite being a valid request, the Responder decided to reject it. The Responder guarantees that it didn't process the request. The reason for the rejection is explained in the Error Data section. Stream ID MUST be > 0.
	ErrCodeRejected ErrCode = 0x00000202
	// The Responder canceled the request but may have started processing it (similar to REJECTED but doesn't guarantee lack of side-effects). Stream ID MUST be > 0.
	ErrCodeCanceled ErrCode = 0x00000203
	// The request is invalid. Stream ID MUST be > 0.
	ErrCodeInvalid ErrCode = 0x00000204
)
