package main

import (
	"context"
	"database/sql"
	"fmt"
	"gatekeeper/internal"
	"log/slog"
	"os"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/labstack/gommon/log"
	"github.com/samber/do"
)

func panicOnErr(err error) {
	if err != nil {
		slog.Error("%s", err)
		os.Exit(1)
	}
}

func main() {
	injector := internal.NewInjector()
	defer injector.Shutdown()

	s := gocron.NewScheduler(time.UTC)

	_, err := s.Every(30).Minutes().Name("DeleteExpiredChallengesJob").Do(DeleteExpiredChallengesJob, injector)
	panicOnErr(err)

	s.RegisterEventListeners(
		gocron.WhenJobReturnsError(func(jobName string, err error) {
			log.Errorf("%s: %s", jobName, err)
		}),
	)

	s.StartBlocking()
}

func DeleteExpiredChallengesJob(i *do.Injector) error {
	db := do.MustInvoke[*sql.DB](i)

	_, err := db.ExecContext(context.Background(), "DELETE FROM challenges WHERE expired_at <= ?", time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to delete expired challenges: %w", err)
	}

	return nil
}
