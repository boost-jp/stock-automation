package boilerrors

import (
	"fmt"

	"github.com/boost-jp/adcast-backend/app/errors"
)

// sqlboiler専用のエラー
// sqlboilerのテンプレートを上書きするとバージョンアップの時のコストが上がるため内部的に置換する

func New(message string) error {
	return errors.Wrap(errors.ErrUnknown, message)
}

func Wrap(err error, message string) error {
	return errors.WrapWithCustom(err, fmt.Errorf("%w: %s", errors.ErrUnknown, message))
}

func Is(err error, target error) bool {
	return errors.Is(err, target)
}

func As(err error, target any) bool {
	return errors.As(err, target)
}
