package migration

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigrations(t *testing.T) {
	testData := map[string]int{}
	curVer := &inmemIDStore{}

	t.Run("native", func(t *testing.T) {
		require := require.New(t)

		native := Begin("native")

		num := 1
		lastGood := native.Next("01",
			func() error {
				testData["migration1"] = num
				num++
				return nil
			},
		).Next("02",
			func() error {
				testData["migration2"] = num
				num++
				return nil
			},
		)

		afterBad := lastGood.Next("03",
			func() error {
				testData["migration3"] = num
				num++
				return errors.New("test migration error")
			},
		).Next("04",
			func() error {
				testData["migration4"] = num
				num++
				return nil
			},
		)

		err := afterBad.Exec(curVer)
		require.Error(err, "Success run migration manager with error migrations")

		lastID := curVer.GetID()
		require.Equal(lastGood.Id(), lastID, "Bad last id in idProducer after migration error")

		require.Equal(1, testData["migration1"], "Bad value after run migration1")
		require.Equal(2, testData["migration2"], "Bad value after run migration2")
		require.Equal(3, testData["migration3"], "Bad value after run migration3")
		require.Empty(testData["migration4"], "Bad data for migration4 - should by empty")

		// Continue with fixed transactions

		num = 3
		fixed := lastGood.Next("03",
			func() error {
				testData["migration3"] = num
				num++
				return nil
			},
		).Next("04",
			func() error {
				testData["migration4"] = num
				num++
				return nil
			},
		)

		err = fixed.Exec(curVer)
		require.NoError(err, "Error when run migration manager")

		require.Equal(1, testData["migration1"], "Bad value after run migration1")
		require.Equal(2, testData["migration2"], "Bad value after run migration2")
		require.Equal(3, testData["migration3"], "Bad value after run migration3")
		require.Equal(4, testData["migration4"], "Bad value after run migration4")
	})

	t.Run("nonnative", func(t *testing.T) {
		require := require.New(t)

		nonnative := Begin("nonnative").
			Next("01",
				func() error {
					testData["migration1"] = 999
					return nil
				},
			)

		err := nonnative.Exec(curVer)
		require.NotEqual(999, testData["migration1"], "nonnative migration is applied")
		require.Error(err)
	})
}