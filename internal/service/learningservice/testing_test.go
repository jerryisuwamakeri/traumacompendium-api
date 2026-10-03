package learningservice

import (
	"time"

	"github.com/traumacompendium/backend/internal/domain/learning"
	"github.com/traumacompendium/backend/internal/testsupport/fakes"
)

// seedTwoLessonLevel builds a level with two lessons (the second has a
// quiz with one question/two options) in the given content fake, and
// returns the level and lesson IDs for convenience.
func seedTwoLessonLevel(content *fakes.LearningContentRepo) (levelID, lesson1ID, lesson2ID, quizID, questionID, correctOptionID, wrongOptionID string) {
	levelID = "level-1"
	lesson1ID = "lesson-1"
	lesson2ID = "lesson-2"
	quizID = "quiz-1"
	questionID = "question-1"
	correctOptionID = "option-correct"
	wrongOptionID = "option-wrong"

	content.Levels[levelID] = learning.Level{ID: levelID, Number: 1, Name: "Beginner"}
	content.Lessons[lesson1ID] = learning.Lesson{ID: lesson1ID, LevelID: levelID, Title: "Lesson 1", OrderIndex: 1}
	content.Lessons[lesson2ID] = learning.Lesson{ID: lesson2ID, LevelID: levelID, Title: "Lesson 2", OrderIndex: 2}
	content.Quizzes[lesson2ID] = learning.Quiz{ID: quizID, LessonID: lesson2ID}
	content.Questions[quizID] = []learning.Question{
		{
			ID: questionID, QuizID: quizID, Text: "2+2?", OrderIndex: 1,
			Options: []learning.Option{
				{ID: correctOptionID, QuestionID: questionID, Text: "4", IsCorrect: true, OrderIndex: 1},
				{ID: wrongOptionID, QuestionID: questionID, Text: "5", IsCorrect: false, OrderIndex: 2},
			},
		},
	}
	content.Badges[levelID] = learning.Badge{ID: "badge-1", LevelID: levelID, Name: "Beginner"}

	return
}

func newTestService() (*Service, *fakes.LearningContentRepo, *fakes.LearningProgressRepo) {
	svc, content, progress, _, _ := newTestServiceWithAudio()
	return svc, content, progress
}

func newTestServiceWithAudio() (*Service, *fakes.LearningContentRepo, *fakes.LearningProgressRepo, *fakes.LearningAudioRepo, *fakes.TTSSynthesizer) {
	svc, content, progress, audio, synth, _, _ := newTestServiceWithCertificates()
	return svc, content, progress, audio, synth
}

// newTestServiceWithCertificates is used by tests that need to exercise
// certificate issuance/lookup directly, in addition to everything
// newTestServiceWithAudio's callers already rely on.
func newTestServiceWithCertificates() (*Service, *fakes.LearningContentRepo, *fakes.LearningProgressRepo, *fakes.LearningAudioRepo, *fakes.TTSSynthesizer, *fakes.CertificateRepo, *fakes.UserRepo) {
	content := fakes.NewLearningContentRepo()
	progress := fakes.NewLearningProgressRepo()
	audio := fakes.NewLearningAudioRepo()
	synth := &fakes.TTSSynthesizer{}
	certificates := fakes.NewCertificateRepo()
	users := fakes.NewUserRepo()
	svc := New(content, progress, audio, certificates, users, synth)
	svc.now = func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }
	return svc, content, progress, audio, synth, certificates, users
}
