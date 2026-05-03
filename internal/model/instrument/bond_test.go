package instrument

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetBondCouponsParams_Validate(t *testing.T) {

	t.Run("valid", func(t *testing.T) {
		now := time.Now()

		p := GetBondCouponsParams{
			From: now,
			To:   now.Add(time.Hour),
		}

		require.NoError(t, p.Validate())
	})

	t.Run("invalid time range", func(t *testing.T) {
		now := time.Now()

		p := GetBondCouponsParams{
			From: now,
			To:   now.Add(-time.Hour),
		}

		require.Error(t, p.Validate())
	})
}
