package gocron

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/marksalpeter/sugar"
)

func TestJob(t *testing.T) {

	// note: we're defining today as the first of the month so we can test an important edge case in the lastRun
	// calculation when a jobs lastRun time occured the previous month from when the job initialized
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), 1, now.Hour(), now.Minute(), now.Second(), 0, now.Location())
	aMinuteAgo := today.Add(-time.Minute)
	aMinuteFromNow := today.Add(time.Minute)
	aMinuteAgoAtTime := fmt.Sprintf("%02d:%02d", aMinuteAgo.Hour(), aMinuteAgo.Minute())
	aMinuteFromNowAtTime := fmt.Sprintf("%02d:%02d", aMinuteFromNow.Hour(), aMinuteFromNow.Minute())
	s := sugar.New(t)

	s.Title("Day")

	s.Assert("`Job.Every(...).Day().At(...)`", func(log sugar.Log) bool {
		// try this with 20 random day intervals
		for i := 20; i > 0; i-- {

			// get a random interval of days [1, 5]
			rand.Seed(time.Now().UnixNano())
			interval := 1 + uint64(rand.Int())%6

			// create and init the job
			job := newJob(interval).Day().At(aMinuteFromNowAtTime)
			job.init(today)

			// jobs last run should be `interval` days from now
			aMinuteFromNowIntervalDaysFromnow := aMinuteFromNow
			if !job.lastRun.Equal(aMinuteFromNow) {
				log("the lastRun did not occur a minute from now days ago")
				log(job.lastRun, aMinuteFromNowIntervalDaysFromnow)
				return false
			}
			//
			// // jobs last run should be `interval` days from now
			// aMinuteAgoIntervalDaysAgo := aMinuteAgo.Add(-1 * Day * time.Duration(interval))
			// if !job.lastRun.Equal(aMinuteAgoIntervalDaysAgo) {
			// 	log("the lastRun did not occur %d days ago", interval)
			// 	log(job.lastRun, aMinuteFromNowIntervalDaysAgo)
			// 	return false
			// }

			// jobs next run is should be today
			if !job.nextRun.Equal(aMinuteFromNow) {
				log("the nextRun will not happen a minute from now")
				log(job.nextRun, aMinuteFromNow)
				return false
			}

			// after run, the nextRun is interval days from the previous nextRun
			job.run()
			aMinutFromNowIntervalDaysAfterNextRun := aMinuteFromNow.Add(Day * time.Duration(interval))
			if !job.nextRun.Equal(aMinutFromNowIntervalDaysAfterNextRun) {
				log("the next nextRun will not happen in %d days", interval)
				log(job.nextRun, aMinutFromNowIntervalDaysAfterNextRun)
				return false
			}

		}

		return true
	})

	s.Assert("`Job.Every(...).Day.At(...)` set to the past", func(log sugar.Log) bool {
		// try this with 20 random day intervals
		for i := 20; i > 0; i-- {

			// get a random interval of days [1, 5]
			rand.Seed(time.Now().UnixNano())
			interval := 1 + uint64(rand.Int())%5

			// create and init the job
			job := newJob(interval).Day().At(aMinuteAgoAtTime)
			job.init(today)

			// jobs last run interval days from tomorrow
			aMinuteAgoIntervalDaysFromTomorrow := aMinuteAgo.Add(Day).Add(-1 * Day * time.Duration(interval))
			if !job.lastRun.Equal(aMinuteAgoIntervalDaysFromTomorrow) {
				log("the lastRun did not %d days from tomorrow", interval)
				log(job.lastRun, aMinuteAgoIntervalDaysFromTomorrow)
				return false
			}

			// jobs next run is tomorrow
			aMinuteAgoTomorrow := aMinuteAgo.Add(Day)
			if !job.nextRun.Equal(aMinuteAgoTomorrow) {
				log("the nextRun will not occur tomorrow")
				log(job.nextRun, aMinuteAgoTomorrow)
				return false
			}

			// after run, the nextRun is interval days from the previous nextRun
			job.run()
			aMinutAgoIntervalDaysAfterNextRun := aMinuteAgoTomorrow.Add(Day * time.Duration(interval))
			if !job.nextRun.Equal(aMinutAgoIntervalDaysAfterNextRun) {
				log("the next nextRun will not happen in %d days", interval)
				log(job.nextRun, aMinutAgoIntervalDaysAfterNextRun)
				return false
			}
		}

		return true
	})

	s.Title("Week")

	s.Assert("`Job.Every(...).Weekday(...).At(...)` set to the past", func(log sugar.Log) bool {
		// try this with 20 random weekdays and week intervals
		for i := 20; i > 0; i-- {

			// get a random interval of weeks [1, 52]
			rand.Seed(time.Now().UnixNano())
			interval := 1 + uint64(rand.Int())%52

			// get a random day of the week that is today or before today
			rand.Seed(time.Now().UnixNano())
			weekday := time.Weekday(rand.Int() % int(today.Weekday()+1))
			durationAfterWeekday := time.Duration(weekday-today.Weekday()) * 24 * time.Hour

			// create and init the job
			job := newJob(interval).Weekday(weekday).At(aMinuteAgoAtTime)
			job.init(today)

			// jobs lastRun was interval weeks ago from next week
			aMinuteAgoIntervalWeeksFromNextWeek := aMinuteAgo.Add(durationAfterWeekday).Add(Week).Add(-1 * Week * time.Duration(interval))
			if !job.lastRun.Equal(aMinuteAgoIntervalWeeksFromNextWeek) {
				log("the lastRun did not occur %d weeks ago", interval+1)
				log(weekday, aMinuteAgoIntervalWeeksFromNextWeek.Weekday(), job.nextRun.Weekday(), job.lastRun, aMinuteAgoIntervalWeeksFromNextWeek)
				return false
			}

			// jobs next run is next week
			aMinuteAgoNextWeek := aMinuteAgo.Add(durationAfterWeekday).Add(Week)
			if !job.nextRun.Equal(aMinuteAgoNextWeek) {
				log("the nextRun will not occur next week")
				log(weekday, aMinuteAgoNextWeek.Weekday(), job.nextRun.Weekday(), job.nextRun, aMinuteAgoNextWeek)
				return false
			}

			// after run, the nextRun is interval weeks from the previous nextRun
			job.run()
			aMinutAgoIntervalWeeksAfterNextRun := aMinuteAgoNextWeek.Add(Week * time.Duration(interval))
			if !job.nextRun.Equal(aMinutAgoIntervalWeeksAfterNextRun) {
				log("the next nextRun will not happen in %d weeks", interval)
				log(job.nextRun, aMinutAgoIntervalWeeksAfterNextRun)
				return false
			}

		}

		return true
	})

	s.Assert("`Job.Every(...).Weekday(...).At(...)` set to the future", func(log sugar.Log) bool {
		// try this with 20 random weekdays and week intervals
		for i := 20; i > 0; i-- {

			// get a random interval of weeks [1, 52]
			rand.Seed(time.Now().UnixNano())
			interval := 1 + uint64(rand.Int())%52

			// get a random day of the week that is today or after today
			rand.Seed(time.Now().UnixNano())
			weekday := time.Weekday(int(today.Weekday()) + rand.Int()%(7-int(today.Weekday())))
			durationUntilWeekday := time.Duration(weekday-today.Weekday()) * 24 * time.Hour

			// create and init the job
			job := newJob(interval).Weekday(weekday).At(aMinuteFromNowAtTime)
			job.init(today)

			// jobs last run was interval weeks ago
			aMinuteFromNowIntervalWeeksAgo := aMinuteFromNow.Add(durationUntilWeekday).Add(-1 * Week * time.Duration(interval))
			if !job.lastRun.Equal(aMinuteFromNowIntervalWeeksAgo) {
				log("the lastRun did not occur %d weeks ago", interval)
				log(weekday, aMinuteFromNowIntervalWeeksAgo.Weekday(), job.nextRun.Weekday(), job.lastRun, aMinuteFromNowIntervalWeeksAgo)
				return false
			}

			// jobs next run is this week
			thisWeekdayAMinuteFromNow := aMinuteFromNow.Add(durationUntilWeekday)
			if !job.nextRun.Equal(thisWeekdayAMinuteFromNow) {
				log("the nextRun will not occur this week")
				log(weekday, thisWeekdayAMinuteFromNow.Weekday(), job.nextRun.Weekday(), job.nextRun, thisWeekdayAMinuteFromNow)
				return false
			}

			// after run, the nextRun is interval weeks from the previous nextRun
			job.run()
			aMinutAgoIntervalWeeksAfterNextRun := thisWeekdayAMinuteFromNow.Add(Week * time.Duration(interval))
			if !job.nextRun.Equal(aMinutAgoIntervalWeeksAfterNextRun) {
				log("the next nextRun will not happen in %d weeks", interval)
				log(job.nextRun, aMinutAgoIntervalWeeksAfterNextRun)
				return false
			}
		}

		return true
	})

	s.Title("Time")

	s.Assert("`Job.Hour()` causes lastRun to be now and nextRun to be `interval` hour(s) from now", func(log sugar.Log) bool {
		// TODO: implement test
		return false
	})

	s.Assert("`Job.Minute()` causes lastRun to be now and nextRun to be `interval` minute(s) from now", func(log sugar.Log) bool {
		// TODO: implement test
		return false
	})

	s.Assert("`Job.Second()` causes lastRun to be now and nextRun to be `interval` second(s) from now", func(log sugar.Log) bool {
		// TODO: implement test
		return false
	})
}

func TestScheduler(t *testing.T) {

	s := sugar.New(t)

	s.Assert("`runPending(...)` runs all pending jobs", func(log sugar.Log) bool {
		// TODO: implement test
		return false
	})

	s.Assert("`Start()`, `IsRunning()` and `Stop()` perform correctly in asynchrnous environments", func(log sugar.Log) bool {
		// TODO: implement test
		return false
	})

	s.Assert("`Start()` triggers runPending(...) every second", func(log sugar.Log) bool {
		// TODO: implement test
		return false
	})

}

// This is a basic test for the issue described here: https://github.com/jasonlvhit/gocron/issues/23
func TestScheduler_Weekdays(t *testing.T) {
	scheduler := NewScheduler()

	job1 := scheduler.Every(1).Monday().At("23:59")
	job2 := scheduler.Every(1).Wednesday().At("23:59")
	job1.Do(task)
	job2.Do(task)
	t.Logf("job1 scheduled for %s", job1.NextScheduledTime())
	t.Logf("job2 scheduled for %s", job2.NextScheduledTime())
	if job1.NextScheduledTime() == job2.NextScheduledTime() {
		t.Fail()
		t.Logf("Two jobs scheduled at the same time on two different weekdays should never run at the same time.[job1: %s; job2: %s]", job1.NextScheduledTime(), job2.NextScheduledTime())
	}
}

// This ensures that if you schedule a job for today's weekday, but the time is already passed, it will be scheduled for
// next week at the requested time.
func TestScheduler_WeekdaysTodayAfter(t *testing.T) {
	scheduler := NewScheduler()

	now := time.Now()
	timeToSchedule := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()-1, 0, 0, time.Local)

	job := callTodaysWeekday(scheduler.Every(1)).At(fmt.Sprintf("%02d:%02d", timeToSchedule.Hour(), timeToSchedule.Minute()))
	job.Do(task)
	t.Logf("job is scheduled for %s", job.NextScheduledTime())
	if job.NextScheduledTime().Weekday() != timeToSchedule.Weekday() {
		t.Fail()
		t.Logf("Job scheduled for current weekday for earlier time, should still be scheduled for current weekday (but next week)")
	}
	nextWeek := time.Date(now.Year(), now.Month(), now.Day()+7, now.Hour(), now.Minute()-1, 0, 0, time.Local)
	if !job.NextScheduledTime().Equal(nextWeek) {
		t.Fail()
		t.Logf("Job should be scheduled for the correct time next week.")
	}
}

// This is to ensure that if you schedule a job for today's weekday, and the time hasn't yet passed, the next run time
// will be scheduled for today.
func TestScheduler_WeekdaysTodayBefore(t *testing.T) {
	scheduler := NewScheduler()

	now := time.Now()
	timeToSchedule := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+1, 0, 0, time.Local)

	job := callTodaysWeekday(scheduler.Every(1)).At(fmt.Sprintf("%02d:%02d", timeToSchedule.Hour(), timeToSchedule.Minute()))
	job.Do(task)
	t.Logf("job is scheduled for %s", job.NextScheduledTime())
	if !job.NextScheduledTime().Equal(timeToSchedule) {
		t.Fail()
		t.Logf("Job should be run today, at the set time.")
	}
}

func Test_formatTime(t *testing.T) {
	tests := []struct {
		name     string
		args     string
		wantHour int
		wantMin  int
		wantErr  bool
	}{
		{
			name:     "normal",
			args:     "16:18",
			wantHour: 16,
			wantMin:  18,
			wantErr:  false,
		},
		{
			name:     "normal",
			args:     "6:18",
			wantHour: 6,
			wantMin:  18,
			wantErr:  false,
		},
		{
			name:     "notnumber",
			args:     "e:18",
			wantHour: 0,
			wantMin:  0,
			wantErr:  true,
		},
		{
			name:     "outofrange",
			args:     "25:18",
			wantHour: 25,
			wantMin:  18,
			wantErr:  true,
		},
		{
			name:     "wrongformat",
			args:     "19:18:17",
			wantHour: 0,
			wantMin:  0,
			wantErr:  true,
		},
		{
			name:     "wrongminute",
			args:     "19:1e",
			wantHour: 19,
			wantMin:  0,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHour, gotMin, err := formatTime(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("formatTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotHour != tt.wantHour {
				t.Errorf("formatTime() gotHour = %v, want %v", gotHour, tt.wantHour)
			}
			if gotMin != tt.wantMin {
				t.Errorf("formatTime() gotMin = %v, want %v", gotMin, tt.wantMin)
			}
		})
	}
}

// utility function for testing the weekday functions *on* the current weekday.
func callTodaysWeekday(job *Job) *Job {
	switch time.Now().Weekday() {
	case 0:
		job.Sunday()
	case 1:
		job.Monday()
	case 2:
		job.Tuesday()
	case 3:
		job.Wednesday()
	case 4:
		job.Thursday()
	case 5:
		job.Friday()
	case 6:
		job.Saturday()
	}
	return job
}
