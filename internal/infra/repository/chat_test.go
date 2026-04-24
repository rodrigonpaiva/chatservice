package repository

import (
	"context"
	"database/sql"
	"reflect"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rodrigonpaiva/fclx/chatservice/internal/domain/entity"
)

func TestChatRepository_CreateChat_ThenFindChatByID_PreservesInitialSystemMessage(t *testing.T) {
	ctx := context.Background()

	dbConn, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/chat_app?parseTime=true")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() {
		_ = dbConn.Close()
	})

	if _, err := dbConn.Exec("DELETE FROM messages"); err != nil {
		t.Fatalf("delete messages: %v", err)
	}

	if _, err := dbConn.Exec("DELETE FROM chats"); err != nil {
		t.Fatalf("delete chats: %v", err)
	}

	repo := NewChatRepositoryMySQL(dbConn)

	model := entity.NewModel("gpt-3.5-turbo", 4096)

	initialMessage, err := entity.NewMessage(
		"system",
		"You are a helpful assistant. Your Name is Alpha.",
		model,
	)
	if err != nil {
		t.Fatalf("create initial message: %v", err)
	}

	chatConfig := &entity.ChatConfig{
		Model:       model,
		Temperature: 0.7,
		TopP:        1,
		N:           1,
		Stop:        []string{},
		MaxTokens:   512,
	}

	createdChat, err := entity.NewChat("user-1", initialMessage, chatConfig)
	if err != nil {
		t.Fatalf("create chat: %v", err)
	}

	if err := repo.CreateChat(ctx, createdChat); err != nil {
		t.Fatalf("persist chat: %v", err)
	}

	loadedChat, err := repo.FindChatByID(ctx, createdChat.ID)
	if err != nil {
		t.Fatalf("find chat by id: %v", err)
	}

	if loadedChat.ID != createdChat.ID {
		t.Fatalf("expected chat ID %q, got %q", createdChat.ID, loadedChat.ID)
	}

	if loadedChat.UserID != createdChat.UserID {
		t.Fatalf("expected user ID %q, got %q", createdChat.UserID, loadedChat.UserID)
	}

	if loadedChat.InitialSystemMessage == nil {
		t.Fatal("expected initial system message to be loaded")
	}

	if loadedChat.InitialSystemMessage.ID != initialMessage.ID {
		t.Fatalf("expected initial message ID %q, got %q", initialMessage.ID, loadedChat.InitialSystemMessage.ID)
	}

	if loadedChat.InitialSystemMessage.Content != initialMessage.Content {
		t.Fatalf("expected initial message content %q, got %q", initialMessage.Content, loadedChat.InitialSystemMessage.Content)
	}

	if loadedChat.InitialSystemMessage.Role != "system" {
		t.Fatalf("expected initial message role %q, got %q", "system", loadedChat.InitialSystemMessage.Role)
	}

	if len(loadedChat.Messages) < 1 {
		t.Fatal("expected at least one message in the loaded chat")
	}

	if loadedChat.Messages[0].Role != "system" {
		t.Fatalf("expected first loaded message role %q, got %q", "system", loadedChat.Messages[0].Role)
	}

	if loadedChat.Messages[0].Content != initialMessage.Content {
		t.Fatalf("expected first loaded message content %q, got %q", initialMessage.Content, loadedChat.Messages[0].Content)
	}

	if !reflect.DeepEqual(loadedChat.Config.Stop, []string{}) {
		t.Fatalf("expected stop config %v, got %v", []string{}, loadedChat.Config.Stop)
	}
}

func TestSerializeStop_PreservesEmptyAndCommaSeparatedContentAsJSON(t *testing.T) {
	stop := []string{"hello, world", "fim"}

	raw, err := serializeStop(stop)
	if err != nil {
		t.Fatalf("serialize stop: %v", err)
	}

	if raw != "[\"hello, world\",\"fim\"]" {
		t.Fatalf("expected JSON encoded stop, got %q", raw)
	}

	decoded, err := deserializeStop(raw)
	if err != nil {
		t.Fatalf("deserialize stop: %v", err)
	}

	if !reflect.DeepEqual(decoded, stop) {
		t.Fatalf("expected decoded stop %v, got %v", stop, decoded)
	}
}

func TestDeserializeStop_BackwardCompatibilityWithLegacyCommaSeparatedFormat(t *testing.T) {
	decoded, err := deserializeStop("A,B")
	if err != nil {
		t.Fatalf("deserialize legacy stop: %v", err)
	}

	expected := []string{"A", "B"}
	if !reflect.DeepEqual(decoded, expected) {
		t.Fatalf("expected decoded stop %v, got %v", expected, decoded)
	}
}

func TestChatRepository_CreateChat_RollsBackWhenInitialMessageInsertFails(t *testing.T) {
	ctx := context.Background()
	dbConn := openTestDB(t)
	cleanTables(t, dbConn)

	repo := NewChatRepositoryMySQL(dbConn)

	model := entity.NewModel("gpt-3.5-turbo", 4096)

	existingMessage, err := entity.NewMessage("system", "existing", model)
	if err != nil {
		t.Fatalf("create existing message: %v", err)
	}

	existingChat, err := entity.NewChat("existing-user", existingMessage, defaultChatConfig(model))
	if err != nil {
		t.Fatalf("create existing chat: %v", err)
	}

	if err := repo.CreateChat(ctx, existingChat); err != nil {
		t.Fatalf("persist existing chat: %v", err)
	}

	conflictingInitialMessage, err := entity.NewMessage("system", "new initial", model)
	if err != nil {
		t.Fatalf("create conflicting message: %v", err)
	}
	conflictingInitialMessage.ID = existingMessage.ID

	conflictingChat, err := entity.NewChat("new-user", conflictingInitialMessage, defaultChatConfig(model))
	if err != nil {
		t.Fatalf("create conflicting chat: %v", err)
	}

	err = repo.CreateChat(ctx, conflictingChat)
	if err == nil {
		t.Fatal("expected create chat to fail when initial message insert conflicts")
	}

	if countRows(t, dbConn, "SELECT COUNT(*) FROM chats WHERE id = ?", conflictingChat.ID) != 0 {
		t.Fatal("expected chat row to be rolled back after message insert failure")
	}
}

func TestChatRepository_SaveChat_RollsBackWhenMessageReinsertFails(t *testing.T) {
	ctx := context.Background()
	dbConn := openTestDB(t)
	cleanTables(t, dbConn)

	repo := NewChatRepositoryMySQL(dbConn)

	model := entity.NewModel("gpt-3.5-turbo", 4096)

	initialMessage, err := entity.NewMessage("system", "initial", model)
	if err != nil {
		t.Fatalf("create initial message: %v", err)
	}

	chat, err := entity.NewChat("user-1", initialMessage, defaultChatConfig(model))
	if err != nil {
		t.Fatalf("create chat: %v", err)
	}

	userMessage, err := entity.NewMessage("user", "hello", model)
	if err != nil {
		t.Fatalf("create user message: %v", err)
	}
	if err := chat.AddMessage(userMessage); err != nil {
		t.Fatalf("add user message: %v", err)
	}

	if err := repo.CreateChat(ctx, chat); err != nil {
		t.Fatalf("persist chat: %v", err)
	}

	if err := repo.SaveChat(ctx, chat); err != nil {
		t.Fatalf("save chat baseline state: %v", err)
	}

	loadedChat, err := repo.FindChatByID(ctx, chat.ID)
	if err != nil {
		t.Fatalf("load chat: %v", err)
	}

	duplicateMessage, err := entity.NewMessage("assistant", "response", model)
	if err != nil {
		t.Fatalf("create duplicate message: %v", err)
	}
	duplicateMessage.ID = loadedChat.Messages[0].ID
	loadedChat.Messages = append(loadedChat.Messages, duplicateMessage)

	err = repo.SaveChat(ctx, loadedChat)
	if err == nil {
		t.Fatal("expected save chat to fail when reinserting duplicate message IDs")
	}

	messageCount := countRows(t, dbConn, "SELECT COUNT(*) FROM messages WHERE chat_id = ?", chat.ID)
	if messageCount != 2 {
		t.Fatalf("expected original messages to remain after rollback, got %d", messageCount)
	}

	reloadedChat, err := repo.FindChatByID(ctx, chat.ID)
	if err != nil {
		t.Fatalf("reload chat after rollback: %v", err)
	}

	if len(reloadedChat.Messages) != 2 {
		t.Fatalf("expected 2 original messages after rollback, got %d", len(reloadedChat.Messages))
	}

	if reloadedChat.Messages[0].ID != loadedChat.Messages[0].ID {
		t.Fatalf("expected first message ID %q after rollback, got %q", loadedChat.Messages[0].ID, reloadedChat.Messages[0].ID)
	}

	if reloadedChat.Messages[1].ID != loadedChat.Messages[1].ID {
		t.Fatalf("expected second message ID %q after rollback, got %q", loadedChat.Messages[1].ID, reloadedChat.Messages[1].ID)
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dbConn, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/chat_app?parseTime=true")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	t.Cleanup(func() {
		_ = dbConn.Close()
	})

	return dbConn
}

func cleanTables(t *testing.T, dbConn *sql.DB) {
	t.Helper()

	if _, err := dbConn.Exec("DELETE FROM messages"); err != nil {
		t.Fatalf("delete messages: %v", err)
	}

	if _, err := dbConn.Exec("DELETE FROM chats"); err != nil {
		t.Fatalf("delete chats: %v", err)
	}
}

func defaultChatConfig(model *entity.Model) *entity.ChatConfig {
	return &entity.ChatConfig{
		Model:       model,
		Temperature: 0.7,
		TopP:        1,
		N:           1,
		Stop:        []string{},
		MaxTokens:   512,
	}
}

func countRows(t *testing.T, dbConn *sql.DB, query string, args ...any) int {
	t.Helper()

	var count int
	if err := dbConn.QueryRow(query, args...).Scan(&count); err != nil {
		t.Fatalf("count rows with query %q: %v", query, err)
	}

	return count
}
