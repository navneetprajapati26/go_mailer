package scheduler

import (
	"fmt"
	"go_mailer/config"
	"go_mailer/mailer"
	"go_mailer/template"
	"log"
	"sync"
	"time"
)

// EmailJob represents a scheduled email job
type EmailJob struct {
	ID           string
	To           string
	Subject      string
	TemplatePath string
	TemplateData template.TemplateData
	SendAt       time.Time
	Status       string // "pending", "sent", "failed"
	Error        error
}

// EmailCallback is a function that is called when an email is sent
type EmailCallback func(successful bool)

// Scheduler manages scheduled email jobs
type Scheduler struct {
	mailClient *mailer.Mailer
	jobs       map[string]*EmailJob
	callbacks  map[string]EmailCallback
	mu         sync.RWMutex
	stopChan   chan struct{}
	wg         sync.WaitGroup
}

// New creates a new Scheduler instance
func New(cfg *config.Config) *Scheduler {
	return &Scheduler{
		mailClient: mailer.New(cfg),
		jobs:       make(map[string]*EmailJob),
		callbacks:  make(map[string]EmailCallback),
		stopChan:   make(chan struct{}),
	}
}

// RegisterCallback registers a callback function for a specific job
func (s *Scheduler) RegisterCallback(jobID string, callback EmailCallback) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.jobs[jobID]
	if !exists {
		log.Printf("Warning: Trying to register callback for non-existent job ID: %s", jobID)
		return
	}

	s.callbacks[jobID] = callback
}

// ScheduleEmail schedules an email to be sent at a specific time
func (s *Scheduler) ScheduleEmail(to, subject, templatePath string, templateData template.TemplateData, sendAt time.Time) (string, error) {
	// Generate a unique ID for the job
	id := fmt.Sprintf("job-%d", time.Now().UnixNano())

	job := &EmailJob{
		ID:           id,
		To:           to,
		Subject:      subject,
		TemplatePath: templatePath,
		TemplateData: templateData,
		SendAt:       sendAt,
		Status:       "pending",
	}

	s.mu.Lock()
	s.jobs[id] = job
	s.mu.Unlock()

	log.Printf("Email scheduled with ID '%s' to be sent at %s", id, sendAt.Format(time.RFC1123))
	return id, nil
}

// GetJob retrieves information about a specific job
func (s *Scheduler) GetJob(id string) (*EmailJob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, exists := s.jobs[id]
	if !exists {
		return nil, fmt.Errorf("job with ID '%s' not found", id)
	}

	return job, nil
}

// ListJobs returns all scheduled jobs
func (s *Scheduler) ListJobs() []*EmailJob {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := make([]*EmailJob, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}

	return jobs
}

// CancelJob cancels a scheduled job
func (s *Scheduler) CancelJob(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[id]
	if !exists {
		return fmt.Errorf("job with ID '%s' not found", id)
	}

	if job.Status != "pending" {
		return fmt.Errorf("job with ID '%s' has already been processed (status: %s)", id, job.Status)
	}

	delete(s.jobs, id)
	log.Printf("Job with ID '%s' has been cancelled", id)
	return nil
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	log.Println("Email scheduler started")

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.processJobs()
			case <-s.stopChan:
				return
			}
		}
	}()
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	log.Println("Stopping email scheduler...")
	close(s.stopChan)
	s.wg.Wait()
	log.Println("Email scheduler stopped")
}

// processJobs processes jobs that are due
func (s *Scheduler) processJobs() {
	now := time.Now()
	var jobsToProcess []*EmailJob

	// First, find jobs that need to be processed
	s.mu.RLock()
	for _, job := range s.jobs {
		if job.Status == "pending" && now.After(job.SendAt) {
			jobsToProcess = append(jobsToProcess, job)
		}
	}
	s.mu.RUnlock()

	// Process each job
	for _, job := range jobsToProcess {
		go func(j *EmailJob) {
			log.Printf("Processing scheduled email job '%s'", j.ID)

			// Send the email
			err := s.mailClient.SendWithTemplate(j.To, j.Subject, j.TemplatePath, j.TemplateData)

			// Update job status
			s.mu.Lock()
			var successful bool
			if err != nil {
				j.Status = "failed"
				j.Error = err
				log.Printf("Failed to send scheduled email '%s': %v", j.ID, err)
				successful = false
			} else {
				j.Status = "sent"
				log.Printf("Scheduled email '%s' sent successfully", j.ID)
				successful = true
			}

			// Get the callback if it exists
			callback, hasCallback := s.callbacks[j.ID]
			s.mu.Unlock()

			// Execute the callback if it exists
			if hasCallback {
				go callback(successful)
			}
		}(job)
	}
}
