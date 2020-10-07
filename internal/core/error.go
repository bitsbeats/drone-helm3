package core

type ErrorKind string

const (
	// PreFail is used if anything before the helm deployment fails
	PreFailErrorKind = "prefail"

	// PostFailErrorKind is used if a postcmd fails
	PostFailErrorKind = "postfail"

	// Failed is used if the deployment failed
	FailedErrorKind = "failed"

	// TestsFailed is used if the deployment is successful but the
	// `helm test` failed and rollback is not specified
	TestFailedErrorKind = "test_failed"

	// RollbackSuccess is used if the deployment was successful, the tests
	// failed and the rollback was successful
	RollbackSuccessErrorKind = "rollback_success"

	// RollbackFailed is used if the deployment was successful the tests
	// failed and the rollback failed also
	RollbackFailedErrorKind = "rollback_failed"
)
