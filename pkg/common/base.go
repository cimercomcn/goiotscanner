package common

type RR uint32

const (
    // RR = Running Result
    RR_PASS                RR = 0
    RR_ERROR_CHECK_ENV     RR = 0x1000
    RR_ERROR_EXTRACTED_BIN RR = 0x2001
    RR_ERROR_ENCRYPTED_BIN RR = 0x2002
)

// global report
var GReport Report
