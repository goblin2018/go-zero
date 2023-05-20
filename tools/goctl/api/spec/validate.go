package spec

import "errors"

var ErrMissingService = errors.New("missing service")

// Validate validates Validate the integrity of the spec.
func (s *ApiSpec) Validate() error {

	return nil
}
