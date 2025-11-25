//go:build integration

package integration_tests

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"AVITOSAMPISHU/internal/domain"
	"AVITOSAMPISHU/internal/infrastructure/database"
	pullrequest_repository "AVITOSAMPISHU/internal/repository/pullrequest_repository"
	reviewer_repository "AVITOSAMPISHU/internal/repository/reviewer_repository"
	team_repository "AVITOSAMPISHU/internal/repository/team_repository"
	user_repository "AVITOSAMPISHU/internal/repository/user_repository"
	pullrequest_service "AVITOSAMPISHU/internal/service/pullrequest_service"
	team_service "AVITOSAMPISHU/internal/service/team_service"
	user_service "AVITOSAMPISHU/internal/service/user_service"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

var (
	testDB          *sql.DB
	teamRepo        *team_repository.TeamStorage
	userRepo        *user_repository.UserRepository
	prRepo          *pullrequest_repository.PullRequestStorage
	prReviewersRepo *reviewer_repository.PrReviewersStorage
)

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	testDB, err = database.NewDB(ctx)
	if err != nil {
		os.Exit(1)
	}
	defer testDB.Close()

	teamRepo = team_repository.NewTeamStorage(testDB)
	userRepo = user_repository.NewUserRepository(testDB)
	prRepo = pullrequest_repository.NewPullRequestStorage(testDB)
	prReviewersRepo = reviewer_repository.NewPrReviewersStorage(testDB)

	if err := cleanupDB(testDB); err != nil {
		os.Exit(1)
	}

	code := m.Run()

	_ = cleanupDB(testDB)

	os.Exit(code)
}

func cleanupDB(db *sql.DB) error {
	queries := []string{
		"DELETE FROM reviewers",
		"DELETE FROM pull_requests",
		"DELETE FROM users",
		"DELETE FROM teams",
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

// TestIntegrationPullRequestFlow проверяет полный флоу работы с PR:
// создание команды, создание пользователей, создание PR, назначение ревьюверов,
// мердж PR с проверкой изменения статуса и заполнения merged_at,
// проверка идемпотентности merge, проверка cascade delete.
func TestIntegrationPullRequestFlow(t *testing.T) {
	ctx := context.Background()

	// Создание команды
	teamName := "Backend"
	team := &domain.Team{
		TeamName: teamName,
		Members: []domain.TeamMember{
			{UserID: "user-alice-1", Username: "Alice", IsActive: true},     // автор
			{UserID: "user-bob-1", Username: "Bob", IsActive: true},         // ревьювер 1
			{UserID: "user-charlie-1", Username: "Charlie", IsActive: true}, // ревьювер 2
			{UserID: "user-david-1", Username: "David", IsActive: false},    // неактивный (не должен быть назначен)
		},
	}

	teamSvc := team_service.NewTeamService(teamRepo, userRepo)
	createdTeam, err := teamSvc.CreateTeam(ctx, team)
	require.NoError(t, err, "команда не создалась")
	require.NotEmpty(t, createdTeam.TeamName, "имя команды должно быть заполнено")

	authorID := "user-alice-1"
	reviewer1ID := "user-bob-1"
	reviewer2ID := "user-charlie-1"
	inactiveUserID := "user-david-1"

	// Создание Pull Request
	prSvc := pullrequest_service.NewPullRequestService(prRepo, prReviewersRepo, userRepo, teamRepo)
	createPRReq := &domain.CreatePullRequestReq{
		PullRequestID:   "pr-feature-1",
		PullRequestName: "Add new feature",
		AuthorID:        authorID,
	}

	pr, err := prSvc.CreatePullRequest(ctx, createPRReq)
	require.NoError(t, err, "PR не создался")
	require.Equal(t, "pr-feature-1", pr.PullRequestID, "ID PR должен совпадать")
	require.Equal(t, domain.PRStatusOpen, pr.Status, "статус должен быть OPEN")
	require.NotNil(t, pr.CreatedAt, "created_at должен быть заполнен")
	require.False(t, pr.CreatedAt.IsZero(), "created_at должен быть заполнен")

	// Проверка, что ревьюверы назначены
	reviewers, err := prReviewersRepo.GetAssignedReviewers(ctx, pr.PullRequestID)
	require.NoError(t, err, "не удалось получить количество ревьюверов")
	require.GreaterOrEqual(t, len(reviewers), 1, "должен быть назначен хотя бы 1 ревьювер")
	require.LessOrEqual(t, len(reviewers), domain.MaxReviewersCount, "не должно быть больше максимального количества ревьюверов")

	// Проверка, что автор НЕ назначен ревьювером
	require.NotContains(t, reviewers, authorID, "автор не должен быть назначен ревьювером")

	// Проверка, что неактивный пользователь НЕ назначен
	require.NotContains(t, reviewers, inactiveUserID, "неактивный пользователь не должен быть назначен ревьювером")

	// Проверка, что активные ревьюверы назначены
	require.Contains(t, reviewers, reviewer1ID, "Bob должен быть в списке ревьюверов")
	require.Contains(t, reviewers, reviewer2ID, "Charlie должен быть в списке ревьюверов")

	// Проверка статуса PR до мерджа
	prBeforeMerge, err := prRepo.GetPullRequestByID(ctx, pr.PullRequestID)
	require.NoError(t, err, "не удалось выполнить запрос статуса PR")
	require.Equal(t, domain.PRStatusOpen, prBeforeMerge.Status, "статус должен быть OPEN")
	require.Nil(t, prBeforeMerge.MergedAt, "merged_at должен быть NULL для открытого PR")

	// Мердж Pull Request
	mergeReq := &domain.MergePullRequestReq{
		PullRequestID: pr.PullRequestID,
	}

	mergedPR, err := prSvc.MergePullRequest(ctx, mergeReq)
	require.NoError(t, err, "PR должен быть смержен")
	require.Equal(t, domain.PRStatusMerged, mergedPR.Status, "статус должен быть MERGED")
	require.NotNil(t, mergedPR.MergedAt, "merged_at должен быть заполнен")
	require.True(t, mergedPR.MergedAt.After(*pr.CreatedAt), "merged_at должен быть позже created_at")

	// Проверка идемпотентности (повторный мердж)
	mergedPR2, err := prSvc.MergePullRequest(ctx, mergeReq)
	require.NoError(t, err, "повторный мердж не должен вызывать ошибку")
	require.Equal(t, domain.PRStatusMerged, mergedPR2.Status, "статус должен остаться MERGED")

	// Тест cascade delete (удаление команды удаляет все связанные данные)
	_, err = testDB.Exec("DELETE FROM teams WHERE team_name = $1", teamName)
	require.NoError(t, err, "команда должна быть удалена")

	// Проверка, что пользователи удалены
	var usersCount int
	err = testDB.QueryRow("SELECT COUNT(*) FROM users WHERE id IN ($1, $2, $3, $4)", authorID, reviewer1ID, reviewer2ID, inactiveUserID).Scan(&usersCount)
	require.NoError(t, err, "не удалось выполнить запрос пользователей")
	require.Equal(t, 0, usersCount, "пользователи должны быть удалены через cascade")

	// Проверка, что PR удален
	var prsCount int
	err = testDB.QueryRow("SELECT COUNT(*) FROM pull_requests WHERE id = $1", pr.PullRequestID).Scan(&prsCount)
	require.NoError(t, err, "должен быть выполнен запрос PR")
	require.Equal(t, 0, prsCount, "PR должен быть удален через cascade")
}

// TestIntegrationTeamCreation проверяет создание команды и получение команды с участниками
func TestIntegrationTeamCreation(t *testing.T) {
	ctx := context.Background()

	teamSvc := team_service.NewTeamService(teamRepo, userRepo)
	team := &domain.Team{
		TeamName: "Frontend",
		Members: []domain.TeamMember{
			{UserID: "user-1", Username: "User1", IsActive: true},
			{UserID: "user-2", Username: "User2", IsActive: true},
			{UserID: "user-3", Username: "User3", IsActive: false},
		},
	}

	createdTeam, err := teamSvc.CreateTeam(ctx, team)
	require.NoError(t, err, "команда не создалась")
	require.Equal(t, "Frontend", createdTeam.TeamName, "имя команды должно совпадать")

	retrievedTeam, err := teamSvc.GetTeam(ctx, "Frontend")
	require.NoError(t, err, "не удалось получить команду")
	require.Equal(t, "Frontend", retrievedTeam.TeamName, "имя команды должно совпадать")
	require.Len(t, retrievedTeam.Members, 3, "должно быть 3 участника")

	// Проверка, что участники сохранены правильно
	userIDs := make(map[string]bool)
	for _, member := range retrievedTeam.Members {
		userIDs[member.UserID] = true
	}
	require.True(t, userIDs["user-1"], "user-1 должен быть в команде")
	require.True(t, userIDs["user-2"], "user-2 должен быть в команде")
	require.True(t, userIDs["user-3"], "user-3 должен быть в команде")
}

// TestIntegrationDeactivateTeamMembers проверяет массовую деактивацию пользователей команды
// с переназначением ревьюверов на открытых PR
func TestIntegrationDeactivateTeamMembers(t *testing.T) {
	ctx := context.Background()

	teamSvc := team_service.NewTeamService(teamRepo, userRepo)
	userSvc := user_service.NewUserService(userRepo, prReviewersRepo, teamRepo)
	prSvc := pullrequest_service.NewPullRequestService(prRepo, prReviewersRepo, userRepo, teamRepo)

	// Создание команды с участниками
	team := &domain.Team{
		TeamName: "DeactivateTest",
		Members: []domain.TeamMember{
			{UserID: "author-1", Username: "Author1", IsActive: true},
			{UserID: "reviewer-1", Username: "Reviewer1", IsActive: true},
			{UserID: "reviewer-2", Username: "Reviewer2", IsActive: true},
			{UserID: "reviewer-3", Username: "Reviewer3", IsActive: true},
		},
	}

	_, err := teamSvc.CreateTeam(ctx, team)
	require.NoError(t, err, "команда не создалась")

	// Создание PR с назначенными ревьюверами
	createPRReq := &domain.CreatePullRequestReq{
		PullRequestID:   "pr-deactivate-1",
		PullRequestName: "Test PR for deactivation",
		AuthorID:        "author-1",
	}

	pr, err := prSvc.CreatePullRequest(ctx, createPRReq)
	require.NoError(t, err, "PR не создался")

	// Проверка, что ревьюверы назначены
	reviewersBefore, err := prReviewersRepo.GetAssignedReviewers(ctx, pr.PullRequestID)
	require.NoError(t, err, "не удалось получить ревьюверов")
	require.Greater(t, len(reviewersBefore), 0, "должен быть хотя бы один ревьювер")

	// Деактивация ревьювера, который назначен на PR
	req := &domain.DeactivateTeamMembersReq{
		TeamName: "DeactivateTest",
		UserIDs:  []string{"reviewer-1"},
	}

	res, err := userSvc.DeactivateTeamMembers(ctx, req)
	require.NoError(t, err, "деактивация не прошла")
	require.Len(t, res.DeactivatedUserIDs, 1, "должен быть деактивирован 1 пользователь")
	require.Equal(t, "reviewer-1", res.DeactivatedUserIDs[0], "reviewer-1 должен быть деактивирован")

	// Проверка, что пользователь деактивирован
	user, err := userRepo.GetUserByID(ctx, "reviewer-1")
	require.NoError(t, err, "не удалось получить пользователя")
	require.False(t, user.IsActive, "пользователь должен быть неактивным")

	// Если был переназначен ревьювер, проверяем это
	if len(res.Reassignments) > 0 {
		require.Equal(t, pr.PullRequestID, res.Reassignments[0].PrID, "PR ID должен совпадать")
		require.Equal(t, "reviewer-1", res.Reassignments[0].OldReviewerID, "старый ревьювер должен быть reviewer-1")
		require.NotEmpty(t, res.Reassignments[0].NewReviewerID, "новый ревьювер должен быть назначен")
		require.NotEqual(t, "reviewer-1", res.Reassignments[0].NewReviewerID, "новый ревьювер не должен быть reviewer-1")
	}
}

// TestIntegrationReassignReviewer проверяет переназначение ревьювера на PR
func TestIntegrationReassignReviewer(t *testing.T) {
	ctx := context.Background()

	teamSvc := team_service.NewTeamService(teamRepo, userRepo)
	prSvc := pullrequest_service.NewPullRequestService(prRepo, prReviewersRepo, userRepo, teamRepo)

	// Создание команды
	team := &domain.Team{
		TeamName: "ReassignTest",
		Members: []domain.TeamMember{
			{UserID: "author-1", Username: "Author1", IsActive: true},
			{UserID: "reviewer-1", Username: "Reviewer1", IsActive: true},
			{UserID: "reviewer-2", Username: "Reviewer2", IsActive: true},
			{UserID: "reviewer-3", Username: "Reviewer3", IsActive: true},
		},
	}

	_, err := teamSvc.CreateTeam(ctx, team)
	require.NoError(t, err, "команда не создалась")

	// Создание PR
	createPRReq := &domain.CreatePullRequestReq{
		PullRequestID:   "pr-reassign-1",
		PullRequestName: "Test PR for reassignment",
		AuthorID:        "author-1",
	}

	pr, err := prSvc.CreatePullRequest(ctx, createPRReq)
	require.NoError(t, err, "PR не создался")

	// Получение текущих ревьюверов
	reviewersBefore, err := prReviewersRepo.GetAssignedReviewers(ctx, pr.PullRequestID)
	require.NoError(t, err, "не удалось получить ревьюверов")
	require.Greater(t, len(reviewersBefore), 0, "должен быть хотя бы один ревьювер")

	oldReviewerID := reviewersBefore[0]

	// Переназначение ревьювера
	reassignReq := &domain.ReassignReviewerReq{
		PullRequestID: pr.PullRequestID,
		OldUserID:     oldReviewerID,
	}

	reassignedPR, newReviewerID, err := prSvc.ReassignReviewer(ctx, reassignReq)
	require.NoError(t, err, "переназначение не прошло")
	require.NotEmpty(t, newReviewerID, "новый ревьювер должен быть назначен")

	// Проверка, что ревьювер изменился
	reviewersAfter, err := prReviewersRepo.GetAssignedReviewers(ctx, pr.PullRequestID)
	require.NoError(t, err, "не удалось получить ревьюверов после переназначения")
	require.NotContains(t, reviewersAfter, oldReviewerID, "старый ревьювер не должен быть в списке")
	require.Equal(t, len(reviewersBefore), len(reviewersAfter), "количество ревьюверов должно остаться прежним")

	// Проверка статуса PR
	require.Equal(t, domain.PRStatusOpen, reassignedPR.Status, "статус должен остаться OPEN")
}

// TestIntegrationGetUserReviews проверяет получение списка PR для ревьювера
func TestIntegrationGetUserReviews(t *testing.T) {
	ctx := context.Background()

	teamSvc := team_service.NewTeamService(teamRepo, userRepo)
	userSvc := user_service.NewUserService(userRepo, prReviewersRepo, teamRepo)
	prSvc := pullrequest_service.NewPullRequestService(prRepo, prReviewersRepo, userRepo, teamRepo)

	// Создание команды
	team := &domain.Team{
		TeamName: "ReviewsTest",
		Members: []domain.TeamMember{
			{UserID: "author-1", Username: "Author1", IsActive: true},
			{UserID: "reviewer-1", Username: "Reviewer1", IsActive: true},
			{UserID: "reviewer-2", Username: "Reviewer2", IsActive: true},
		},
	}

	_, err := teamSvc.CreateTeam(ctx, team)
	require.NoError(t, err, "команда не создалась")

	// Создание нескольких PR
	pr1Req := &domain.CreatePullRequestReq{
		PullRequestID:   "pr-reviews-1",
		PullRequestName: "PR 1",
		AuthorID:        "author-1",
	}
	pr1, err := prSvc.CreatePullRequest(ctx, pr1Req)
	require.NoError(t, err, "PR 1 не создался")

	pr2Req := &domain.CreatePullRequestReq{
		PullRequestID:   "pr-reviews-2",
		PullRequestName: "PR 2",
		AuthorID:        "author-1",
	}
	pr2, err := prSvc.CreatePullRequest(ctx, pr2Req)
	require.NoError(t, err, "PR 2 не создался")

	// Получение списка PR для ревьювера
	reviews, err := userSvc.GetUserReviews(ctx, "reviewer-1")
	require.NoError(t, err, "не удалось получить список PR")
	require.GreaterOrEqual(t, len(reviews), 1, "должен быть хотя бы один PR")

	// Проверка, что PR в списке
	prIDs := make(map[string]bool)
	for _, review := range reviews {
		prIDs[review.PullRequestID] = true
	}
	require.True(t, prIDs[pr1.PullRequestID] || prIDs[pr2.PullRequestID], "хотя бы один из PR должен быть в списке")
}

// TestIntegrationSetUserIsActive проверяет изменение статуса активности пользователя
func TestIntegrationSetUserIsActive(t *testing.T) {
	ctx := context.Background()

	teamSvc := team_service.NewTeamService(teamRepo, userRepo)
	userSvc := user_service.NewUserService(userRepo, prReviewersRepo, teamRepo)

	// Создание команды с пользователем
	team := &domain.Team{
		TeamName: "ActiveTest",
		Members: []domain.TeamMember{
			{UserID: "user-1", Username: "User1", IsActive: true},
		},
	}

	_, err := teamSvc.CreateTeam(ctx, team)
	require.NoError(t, err, "команда не создалась")

	// Деактивация пользователя
	setActiveReq := &domain.SetIsActiveRequest{
		UserID:   "user-1",
		IsActive: false,
	}

	user, err := userSvc.SetIsActive(ctx, setActiveReq)
	require.NoError(t, err, "не удалось изменить статус")
	require.False(t, user.IsActive, "пользователь должен быть неактивным")

	// Активация пользователя
	setActiveReq.IsActive = true
	user, err = userSvc.SetIsActive(ctx, setActiveReq)
	require.NoError(t, err, "не удалось изменить статус")
	require.True(t, user.IsActive, "пользователь должен быть активным")
}

// TestIntegrationPRWithInsufficientReviewers проверяет создание PR с недостаточным количеством ревьюверов
func TestIntegrationPRWithInsufficientReviewers(t *testing.T) {
	ctx := context.Background()

	teamSvc := team_service.NewTeamService(teamRepo, userRepo)
	prSvc := pullrequest_service.NewPullRequestService(prRepo, prReviewersRepo, userRepo, teamRepo)

	// Создание команды с минимальным количеством участников (только автор)
	team := &domain.Team{
		TeamName: "MinimalTeam",
		Members: []domain.TeamMember{
			{UserID: "author-1", Username: "Author1", IsActive: true},
		},
	}

	_, err := teamSvc.CreateTeam(ctx, team)
	require.NoError(t, err, "команда не создалась")

	// Создание PR (не должно быть достаточно ревьюверов)
	createPRReq := &domain.CreatePullRequestReq{
		PullRequestID:   "pr-insufficient-1",
		PullRequestName: "PR with insufficient reviewers",
		AuthorID:        "author-1",
	}

	pr, err := prSvc.CreatePullRequest(ctx, createPRReq)
	require.NoError(t, err, "PR должен создаться даже без ревьюверов")
	require.NotNil(t, pr.NeedMoreReviewers, "флаг need_more_reviewers должен быть установлен")
	require.True(t, *pr.NeedMoreReviewers, "need_more_reviewers должен быть true")
}

// TestIntegrationMergeAlreadyMergedPR проверяет идемпотентность мерджа уже смерженного PR
func TestIntegrationMergeAlreadyMergedPR(t *testing.T) {
	ctx := context.Background()

	teamSvc := team_service.NewTeamService(teamRepo, userRepo)
	prSvc := pullrequest_service.NewPullRequestService(prRepo, prReviewersRepo, userRepo, teamRepo)

	// Создание команды
	team := &domain.Team{
		TeamName: "IdempotentTest",
		Members: []domain.TeamMember{
			{UserID: "author-1", Username: "Author1", IsActive: true},
			{UserID: "reviewer-1", Username: "Reviewer1", IsActive: true},
		},
	}

	_, err := teamSvc.CreateTeam(ctx, team)
	require.NoError(t, err, "команда не создалась")

	// Создание и мердж PR
	createPRReq := &domain.CreatePullRequestReq{
		PullRequestID:   "pr-idempotent-1",
		PullRequestName: "PR for idempotent test",
		AuthorID:        "author-1",
	}

	pr, err := prSvc.CreatePullRequest(ctx, createPRReq)
	require.NoError(t, err, "PR не создался")

	mergeReq := &domain.MergePullRequestReq{
		PullRequestID: pr.PullRequestID,
	}

	// Первый мердж
	mergedPR1, err := prSvc.MergePullRequest(ctx, mergeReq)
	require.NoError(t, err, "первый мердж не прошел")
	require.Equal(t, domain.PRStatusMerged, mergedPR1.Status, "статус должен быть MERGED")
	require.NotNil(t, mergedPR1.MergedAt, "merged_at должен быть заполнен")

	firstMergedAt := mergedPR1.MergedAt

	// Второй мердж (идемпотентный)
	mergedPR2, err := prSvc.MergePullRequest(ctx, mergeReq)
	require.NoError(t, err, "второй мердж не должен вызывать ошибку")
	require.Equal(t, domain.PRStatusMerged, mergedPR2.Status, "статус должен остаться MERGED")
	require.NotNil(t, mergedPR2.MergedAt, "merged_at должен остаться заполненным")

	// Проверка, что merged_at не изменился (или изменился минимально из-за времени выполнения)
	require.True(t, mergedPR2.MergedAt.Equal(*firstMergedAt) || mergedPR2.MergedAt.After(*firstMergedAt), "merged_at не должен измениться значительно")
}
