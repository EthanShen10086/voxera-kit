package errors_test

import (
	stderrors "errors"
	"strings"
	"testing"

	"github.com/EthanShen10086/voxera-kit/errors"
)

func TestAppError_Error(t *testing.T) {
	t.Run("without cause", func(t *testing.T) {
		err := errors.New(errors.NotFound, "missing")
		if got := err.Error(); got != "[3] missing" {
			t.Fatalf("Error() = %q", got)
		}
	})
	t.Run("with cause", func(t *testing.T) {
		cause := stderrors.New("root")
		err := errors.Wrap(errors.Internal, "failed", cause)
		if !strings.Contains(err.Error(), "failed") || !strings.Contains(err.Error(), "root") {
			t.Fatalf("Error() = %q", err.Error())
		}
	})
}

func TestAppError_Unwrap(t *testing.T) {
	cause := stderrors.New("root")
	err := errors.Wrap(errors.Internal, "wrap", cause)
	if !stderrors.Is(err, cause) {
		t.Fatal("Unwrap should expose cause")
	}
}

func TestConstructors(t *testing.T) {
	if errors.Newf(errors.InvalidArgument, "bad %s", "input").Message != "bad input" {
		t.Fatal("Newf message")
	}
	if errors.Wrapf(errors.Unavailable, stderrors.New("x"), "retry %d", 1).Cause == nil {
		t.Fatal("Wrapf cause")
	}
}

func TestCode(t *testing.T) {
	if errors.Code(nil) != errors.OK {
		t.Fatal("nil -> OK")
	}
	if errors.Code(stderrors.New("plain")) != errors.Unknown {
		t.Fatal("plain -> Unknown")
	}
	if errors.Code(errors.New(errors.Canceled, "canceled")) != errors.Canceled {
		t.Fatal("AppError code")
	}
}

func TestPredicates(t *testing.T) {
	cases := []struct {
		name string
		fn   func(error) bool
		code errors.ErrorCode
	}{
		{"IsNotFound", errors.IsNotFound, errors.NotFound},
		{"IsUnauthorized", errors.IsUnauthorized, errors.Unauthenticated},
		{"IsPermissionDenied", errors.IsPermissionDenied, errors.PermissionDenied},
		{"IsInvalidArgument", errors.IsInvalidArgument, errors.InvalidArgument},
		{"IsInternal", errors.IsInternal, errors.Internal},
		{"IsAlreadyExists", errors.IsAlreadyExists, errors.AlreadyExists},
		{"IsUnavailable", errors.IsUnavailable, errors.Unavailable},
		{"IsDeadlineExceeded", errors.IsDeadlineExceeded, errors.DeadlineExceeded},
		{"IsCanceled", errors.IsCanceled, errors.Canceled},
		{"IsUnimplemented", errors.IsUnimplemented, errors.Unimplemented},
	}
	for _, tc := range cases {
		t.Run(tc.name+" true", func(t *testing.T) {
			if !tc.fn(errors.New(tc.code, "x")) {
				t.Fatalf("%s should be true", tc.name)
			}
		})
		t.Run(tc.name+" false", func(t *testing.T) {
			if tc.fn(errors.New(errors.Unknown, "x")) {
				t.Fatalf("%s should be false", tc.name)
			}
		})
	}
}
