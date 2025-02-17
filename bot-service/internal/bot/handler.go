package bot

import (
	"context"
	"fmt"
	pbScheduler "github.com/Task-bot/bot-service/internal/generated/scheduler"
	"github.com/Task-bot/bot-service/internal/services"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Task-bot/bot-service/internal/generated/notification"
	"github.com/Task-bot/bot-service/internal/generated/task"
	"github.com/Task-bot/bot-service/internal/generated/user"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func SendTelegramNotification(bot *tgbot.BotAPI, userId int64, message string) error {
	msg := tgbot.NewMessage(userId, message)
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
		return err
	}

	return nil

}

func MakeButtons() tgbot.InlineKeyboardMarkup {
	return tgbot.NewInlineKeyboardMarkup(
		[]tgbot.InlineKeyboardButton{tgbot.NewInlineKeyboardButtonData("Добавить задачу", "add_task")},
		[]tgbot.InlineKeyboardButton{tgbot.NewInlineKeyboardButtonData("Посмотреть все задачи", "show_tasks")},
		[]tgbot.InlineKeyboardButton{tgbot.NewInlineKeyboardButtonData("Построить план выполнения", "plan")},
		[]tgbot.InlineKeyboardButton{tgbot.NewInlineKeyboardButtonData("Обновить прогресс задачи", "update_progress")},
		[]tgbot.InlineKeyboardButton{tgbot.NewInlineKeyboardButtonData("Удалить задачу", "delete_task")},
	)
}

func CreateTask(bot *tgbot.BotAPI, message *tgbot.Message, taskClient task.TaskServiceClient, userClient user.UserServiceClient, notificationClient notification.NotificationServiceClient) bool {
	tgUserId := message.From.ID
	text := message.Text
	state, exists := userTaskStates[tgUserId]
	if exists {
		switch state.Step {
		case 1:
			state.TaskName = text
			state.Step = 2
			userTaskStates[tgUserId] = state

			inlineKeyboard := tgbot.NewInlineKeyboardMarkup([]tgbot.InlineKeyboardButton{
				tgbot.NewInlineKeyboardButtonData("Отменить", "cancel_task"),
			})

			msg := tgbot.NewMessage(message.Chat.ID, "Укажите дедлайн задачи (формат: ГГГГ-ММ-ДД ЧЧ:ММ):")
			msg.ReplyMarkup = inlineKeyboard
			_, errSend := bot.Send(msg)
			if errSend != nil {
				log.Printf("Failed to send message: %v", errSend)
			}
			return true

		case 2:
			loc, _ := time.LoadLocation("Europe/Moscow")
			deadline, err := time.ParseInLocation("2006-01-02 15:04", text, loc)
			if err != nil || deadline.Before(time.Now().In(loc)) {
				errorMsg := "Неверный формат даты, попробуйте еще раз"
				if err == nil {
					errorMsg = "Дедлайн должен быть позже текущего времени. Введи еще раз"
				}

				msg := tgbot.NewMessage(message.Chat.ID, errorMsg)
				inlineKeyboard := tgbot.NewInlineKeyboardMarkup([]tgbot.InlineKeyboardButton{
					tgbot.NewInlineKeyboardButtonData("Отменить", "cancel_task"),
				})
				msg.ReplyMarkup = inlineKeyboard
				_, errSend := bot.Send(msg)
				if errSend != nil {
					log.Printf("Failed to send message: %v", errSend)
				}
				return true
			}

			state.Deadline = deadline.UTC()
			state.Step = 3
			userTaskStates[tgUserId] = state

			inlineKeyboard := tgbot.NewInlineKeyboardMarkup([]tgbot.InlineKeyboardButton{
				tgbot.NewInlineKeyboardButtonData("Отменить", "cancel_task"),
			})

			msg := tgbot.NewMessage(message.Chat.ID, "Какой приоритет у этой задачи (1-5, где 1 - слабый, 5 - высокий)?")
			msg.ReplyMarkup = inlineKeyboard
			_, errSend := bot.Send(msg)
			if errSend != nil {
				log.Printf("Failed to send message: %v", errSend)
			}
			return true

		case 3:
			priority, err := strconv.Atoi(text)
			if err != nil || priority < 1 || priority > 5 {
				msg := tgbot.NewMessage(message.Chat.ID, "Приоритет должен быть числом от 1 до 5. Повторите попытку:")
				inlineKeyboard := tgbot.NewInlineKeyboardMarkup([]tgbot.InlineKeyboardButton{
					tgbot.NewInlineKeyboardButtonData("Отменить", "cancel_task"),
				})
				msg.ReplyMarkup = inlineKeyboard
				_, errSend := bot.Send(msg)
				if errSend != nil {
					log.Printf("Failed to send message: %v", errSend)
				}
				return true
			}

			state.Priority = priority
			state.Step = 4
			userTaskStates[tgUserId] = state
			tgUserIdBytes := []byte(strconv.FormatInt(tgUserId, 10))

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			req := &user.GetUserRequest{TgUserId: tgUserIdBytes}
			userGet, err := userClient.GetUser(ctx, req)
			if err != nil {
				log.Printf("Can't get user on task creating: %v", err)
				msg := tgbot.NewMessage(message.Chat.ID, "Ошибка при получении данных пользователя.")
				_, errSend := bot.Send(msg)
				if errSend != nil {
					log.Printf("Failed to send message: %v", errSend)
					return true
				}
				return true
			}
			taskRequest := &task.CreateTaskRequest{
				UserId:   userGet.Id,
				TaskText: state.TaskName,
				Deadline: timestamppb.New(state.Deadline),
				Priority: int32(state.Priority),
			}

			ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, err = taskClient.CreateTask(ctx, taskRequest)
			if err != nil {
				msg := tgbot.NewMessage(message.Chat.ID, "Не удалось создать задачу, попробуйте снова.")
				inlineKeyboard := tgbot.NewInlineKeyboardMarkup([]tgbot.InlineKeyboardButton{
					tgbot.NewInlineKeyboardButtonData("Отменить", "cancel_task"),
				})
				state.Step = 1
				msg.ReplyMarkup = inlineKeyboard
				_, errSend := bot.Send(msg)
				if errSend != nil {
					log.Printf("Failed to send message: %v", errSend)
				}
				return true
			}
			notificationTimes := []time.Duration{
				24 * time.Hour,
				12 * time.Hour,
				3 * time.Hour,
				1 * time.Hour,
				10 * time.Minute,
			}

			now := time.Now().UTC()

			for _, nt := range notificationTimes {
				notifyTime := state.Deadline.Add(-nt)
				if notifyTime.After(now) {
					tgUserID, err := strconv.ParseInt(string(tgUserIdBytes), 10, 64)
					location, err := time.LoadLocation("Europe/Moscow")
					if err != nil {
						log.Println("Error loading location:", err)
					}
					if err != nil {
						log.Fatalf("Ошибка преобразования: %v", err)
					}
					notificationReq := &notification.Notification{
						UserId:     tgUserID,
						Message:    fmt.Sprintf("Напоминание: %s. Дедлайн: %s", state.TaskName, state.Deadline.In(location).Format("02-01-2006 15:04")),
						NotifyTime: timestamppb.New(notifyTime),
					}

					ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					_, err = notificationClient.CreateNotification(ctx, notificationReq)
					if err != nil {
						log.Printf("Не удалось создать уведомление: %v", err)
					}
				}
			}

			inlineKeyboard := MakeButtons()
			msg := tgbot.NewMessage(message.Chat.ID, "Задача успешно добавлена!")
			msg.ReplyMarkup = inlineKeyboard
			_, errSend := bot.Send(msg)
			if errSend != nil {
				log.Printf("Failed to send message: %v", errSend)
			}

			delete(userTaskStates, tgUserId)
			return true
		}
	}
	return false
}

func UpdateTask(bot *tgbot.BotAPI, message *tgbot.Message, taskClient task.TaskServiceClient) bool {
	if state, exists := userTaskProgressUpdateStates[message.From.ID]; exists && state.Step == 1 {
		progress, err := strconv.Atoi(message.Text)
		if err != nil || progress < 0 || progress > 100 {
			_, errSend := bot.Send(tgbot.NewMessage(message.Chat.ID, "Некорректный ввод. Введите число от 0 до 100:"))
			if errSend != nil {
				log.Printf("Failed to send message: %v", errSend)
			}
			return true
		}

		updateReq := &task.UpdateTaskStatusRequest{
			TaskId:   state.TaskId,
			Progress: int32(progress),
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err = taskClient.UpdateTaskStatus(ctx, updateReq)
		if err != nil {
			log.Printf("Failed to update progress: %v", err)
			_, errSend := bot.Send(tgbot.NewMessage(message.Chat.ID, "Ошибка обновления прогресса."))
			if errSend != nil {
				log.Printf("Failed to send message: %v", errSend)
			}
			return true
		}
		inlineKeyboard := MakeButtons()

		msg := tgbot.NewMessage(message.Chat.ID, "Прогресс успешно обновлён!")
		msg.ReplyMarkup = inlineKeyboard
		_, errSend := bot.Send(msg)
		if errSend != nil {
			log.Printf("Failed to send message: %v", errSend)
		}
		delete(userTaskProgressUpdateStates, message.From.ID)
		return true
	}
	return false
}

func handleMessage(bot *tgbot.BotAPI, message *tgbot.Message, registry *services.ServiceRegistry) {
	tgUserId := message.From.ID
	text := message.Text

	tgUserIdBytes := []byte(strconv.FormatInt(tgUserId, 10))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &user.GetUserRequest{TgUserId: tgUserIdBytes}
	service, ok := registry.GetService("user")
	if !ok {
		log.Printf("Failed to get user service")
		return
	}
	userService, ok := service.(*services.UserServiceClient)
	if !ok {
		log.Fatal("Ошибка приведения типа")
	}
	service, ok = registry.GetService("task")
	if !ok {
		log.Printf("Failed to get user service")
		return
	}
	taskService, ok := service.(*services.TaskServiceClient)
	if !ok {
		log.Fatal("Ошибка приведения типа")
	}

	service, ok = registry.GetService("notification")
	if !ok {
		log.Printf("Failed to get user service")
		return
	}
	notificationService, ok := service.(*services.NotificationServiceClient)
	if !ok {
		log.Fatal("Ошибка приведения типа")
	}
	userGet, err := userService.Client.GetUser(ctx, req)
	if err != nil {
		newUser := &user.CreateUserRequest{
			TgUserId: tgUserIdBytes,
			Username: message.From.UserName,
		}
		_, errCreate := userService.Client.CreateUser(ctx, newUser)
		if errCreate != nil {
			log.Printf("Failed to create user: %v", errCreate)
			return
		}

		log.Printf("User created: %v", newUser)
	} else {
		log.Printf("User found: %v", userGet)
	}

	res := CreateTask(bot, message, taskService.Client, userService.Client, notificationService.Client)
	if res {
		return
	}
	res = UpdateTask(bot, message, taskService.Client)
	if res {
		return
	}
	if text == "/start" {
		msg := tgbot.NewMessage(message.Chat.ID, fmt.Sprintf(
			"👋 Привет, %s! Добро пожаловать в *Менеджер задач* 📌\n\n"+
				"✨ Здесь ты можешь:\n"+
				"✅ Создавать задачи с дедлайнами 🕒\n"+
				"🔥 Устанавливать приоритет 📊\n"+
				"🤖 Получать напоминания о сроках ⏳\n"+
				"📅 Планировать выполнение без стресса!\n\n"+
				"🚀 Давай начнем! Используй кнопки ниже ⬇️",
			message.From.FirstName,
		))
		msg.ParseMode = "Markdown"

		_, errSend := bot.Send(msg)
		if errSend != nil {
			log.Printf("Failed to send message after /start: %v", errSend)
		}
		inlineKeyboard := MakeButtons()
		msg = tgbot.NewMessage(message.Chat.ID, "Выбери действие")
		msg.ReplyMarkup = inlineKeyboard
		_, errSend = bot.Send(msg)
		if errSend != nil {
			log.Printf("Failed to send message with buttons: %v", errSend)
		}
		return
	} else {
		msg := tgbot.NewMessage(message.Chat.ID, "Неизвестный запрос или команда, пожалуйста, введите /start")
		_, errSend := bot.Send(msg)
		if errSend != nil {
			log.Printf("Failed to send message: %v", errSend)
		}
		return
	}
}

func handleCallback(bot *tgbot.BotAPI, callback *tgbot.CallbackQuery, registry *services.ServiceRegistry) {
	tgUserId := callback.From.ID
	service, ok := registry.GetService("user")
	if !ok {
		log.Printf("Failed to get user service")
		return
	}
	userService, ok := service.(*services.UserServiceClient)
	if !ok {
		log.Fatal("Ошибка приведения типа")
	}
	service, ok = registry.GetService("task")
	if !ok {
		log.Printf("Failed to get user service")
		return
	}
	taskService, ok := service.(*services.TaskServiceClient)
	if !ok {
		log.Fatal("Ошибка приведения типа")
	}
	switch callback.Data {
	case "add_task":
		userTaskStates[tgUserId] = &TaskCreationState{Step: 1}
		msg := tgbot.NewMessage(callback.Message.Chat.ID, "Введите название задачи:")
		_, errSend := bot.Send(msg)
		if errSend != nil {
			log.Printf("Failed to send message: %v", errSend)
		}
		return
	case "cancel_task":
		delete(userTaskStates, tgUserId)
		msg := tgbot.NewMessage(callback.Message.Chat.ID, "Процесс создания задачи был отменен.")
		_, errSend := bot.Send(msg)
		if errSend != nil {
			log.Printf("Failed to send message: %v", errSend)
		}
		return
	case "show_tasks":
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		tgUserIdBytes := []byte(strconv.FormatInt(tgUserId, 10))

		req := &user.GetUserRequest{TgUserId: tgUserIdBytes}
		userGet, err := userService.Client.GetUser(ctx, req)
		if err != nil {
			log.Printf("Can't get user on show tasks: %v", err)
			msg := tgbot.NewMessage(callback.Message.Chat.ID, "Ошибка при получении данных пользователя.")
			_, errSend := bot.Send(msg)
			if errSend != nil {
				log.Printf("Failed to send message: %v", errSend)
				return
			}
			return
		}
		taskReq := &task.GetTaskRequest{UserId: userGet.Id}
		taskResponse, err := taskService.Client.GetTasks(ctx, taskReq)
		if err != nil {
			log.Printf("Failed to fetch tasks: %v", err)
			msg := tgbot.NewMessage(callback.Message.Chat.ID, "Ошибка при получении списка задач.")
			_, errSend := bot.Send(msg)
			if errSend != nil {
				log.Printf("Failed to send message: %v", errSend)
				return
			}
			return
		}
		if len(taskResponse.Tasks) == 0 {
			msg := tgbot.NewMessage(callback.Message.Chat.ID, "У вас пока нет задач")
			_, errSend := bot.Send(msg)
			if errSend != nil {
				log.Printf("Failed to send message: %v", errSend)
				return
			}
			return
		}

		var taskList string
		loc, _ := time.LoadLocation("Europe/Moscow")

		for i, t := range taskResponse.Tasks {
			emoji := getTaskProgressEmoji(t.Progress)
			deadlineTime := t.Deadline.AsTime().In(loc)
			overdueMsg := ""

			if deadlineTime.Before(time.Now().In(loc)) {
				if t.Progress < 100 {
					overdueMsg = " ❌ Дедлайн просрочен"
				} else {
					overdueMsg = " Молодец! Ты справился с этим делом 🎉"
				}
			}

			taskList += fmt.Sprintf("%d. %s. Дедлайн: [%s] %d%% %s%s\n",
				i+1, t.TaskText, deadlineTime.Format("02.01.2006 15:04"), t.Progress, emoji, overdueMsg)
		}

		msg := tgbot.NewMessage(callback.Message.Chat.ID, taskList)
		_, errSend := bot.Send(msg)
		if errSend != nil {
			log.Printf("Failed to send message: %v", errSend)
			return
		}
		return
	case "plan":
		service, ok = registry.GetService("scheduler")
		if !ok {
			log.Printf("Failed to get user service")
			return
		}
		schedulerService, ok := service.(*services.SchedulerServiceClient)
		if !ok {
			log.Fatal("Ошибка приведения типа")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		tgUserIdBytes := []byte(strconv.FormatInt(tgUserId, 10))
		req := &user.GetUserRequest{TgUserId: tgUserIdBytes}
		userGet, err := userService.Client.GetUser(ctx, req)
		if err != nil {
			log.Printf("Failed to get user on plan stage: %v", err)
		}
		planResp, err := schedulerService.Client.CalculateOptimalPlan(ctx, &pbScheduler.CalculatePlanRequest{
			UserId: userGet.Id,
		})
		if err != nil {
			log.Printf("Ошибка при получении плана: %v", err)
			msg := tgbot.NewMessage(callback.Message.Chat.ID, "Не удалось получить план выполнения задач.")
			_, errSend := bot.Send(msg)
			if errSend != nil {
				log.Printf("Failed to send message: %v", errSend)
				return
			}
			return
		}

		if len(planResp.Tasks) == 0 {
			msg := tgbot.NewMessage(callback.Message.Chat.ID, "У вас нет актуальных задач.")
			_, errSend := bot.Send(msg)
			if errSend != nil {
				log.Printf("Failed to send message: %v", errSend)
				return
			}
			return
		}

		var response string
		for i, oneTask := range planResp.Tasks {
			response += fmt.Sprintf("%d. %s (Приоритет: %d, Дедлайн: %s, Прогресс: %d%%)\n",
				i+1, oneTask.TaskText, oneTask.Priority, oneTask.Deadline.AsTime().Format("2006-01-02"), oneTask.Progress)
		}

		msg := tgbot.NewMessage(callback.Message.Chat.ID, "Оптимальный план:\n"+response)
		_, errSend := bot.Send(msg)
		if errSend != nil {
			log.Printf("Failed to send message: %v", errSend)
			return
		}
		return

	case "update_progress":
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		tgUserIdBytes := []byte(strconv.FormatInt(tgUserId, 10))
		req := &user.GetUserRequest{TgUserId: tgUserIdBytes}
		userGet, err := userService.Client.GetUser(ctx, req)
		if err != nil {
			log.Printf("Can't get user on show tasks: %v", err)
			msg := tgbot.NewMessage(callback.Message.Chat.ID, "Ошибка при получении данных пользователя.")
			_, errSend := bot.Send(msg)
			if errSend != nil {
				log.Printf("Failed to send message: %v", errSend)
				return
			}
			return
		}
		tasksReq := &task.GetTaskRequest{UserId: userGet.Id}
		tasksResp, err := taskService.Client.GetTasks(ctx, tasksReq)
		if err != nil {
			log.Printf("Failed to fetch tasks: %v", err)
			_, errSend := bot.Send(tgbot.NewMessage(callback.Message.Chat.ID, "Ошибка при получении списка задач."))
			if errSend != nil {
				log.Printf("Failed to send message: %v", errSend)
				return
			}
			return
		}

		if len(tasksResp.Tasks) == 0 {
			_, errSend := bot.Send(tgbot.NewMessage(callback.Message.Chat.ID, "У вас нет активных задач."))
			if errSend != nil {
				log.Printf("Failed to send message: %v", errSend)
				return
			}
			return
		}

		timezone, err := time.LoadLocation("Europe/Moscow")
		if err != nil {
			log.Fatalf("Ошибка загрузки часового пояса: %v", err)
		}
		keyboard := tgbot.NewInlineKeyboardMarkup()
		now := time.Now().In(timezone)

		for _, t := range tasksResp.Tasks {
			if t.Deadline != nil {
				deadline := t.Deadline.AsTime().In(timezone)
				if deadline.Before(now) {
					continue
				}
			}
			if t.Progress == 100 {
				continue
			}
			button := tgbot.NewInlineKeyboardButtonData(t.TaskText, "task_"+strconv.FormatInt(t.TaskId, 10))
			keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbot.InlineKeyboardButton{button})
		}
		if len(keyboard.InlineKeyboard) == 0 {
			_, errSend := bot.Send(tgbot.NewMessage(callback.Message.Chat.ID, "У вас нет активных задач."))
			if errSend != nil {
				log.Printf("Failed to send message: %v", errSend)
				return
			}
			return
		}
		msg := tgbot.NewMessage(callback.Message.Chat.ID, "Выберите задачу для обновления прогресса:")
		msg.ReplyMarkup = keyboard
		_, errSend := bot.Send(msg)
		if errSend != nil {
			log.Printf("Failed to send message: %v", errSend)
			return
		}
		return
	case "delete_task":
		tgUserIdBytes := []byte(strconv.FormatInt(tgUserId, 10))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		userReq := &user.GetUserRequest{TgUserId: tgUserIdBytes}
		userGet, err := userService.Client.GetUser(ctx, userReq)
		if err != nil {
			log.Printf("Error fetching user: %v", err)
			_, errSend := bot.Send(tgbot.NewMessage(callback.Message.Chat.ID, "Ошибка: не удалось получить пользователя"))
			if errSend != nil {
				log.Printf("Failed to send message: %v", errSend)
				return
			}
			return
		}

		taskReq := &task.GetTaskRequest{UserId: userGet.Id}
		taskList, err := taskService.Client.GetTasks(ctx, taskReq)
		if err != nil {
			log.Printf("Error fetching tasks: %v", err)
			_, errSend := bot.Send(tgbot.NewMessage(callback.Message.Chat.ID, "Ошибка: не удалось получить список задач"))
			if errSend != nil {
				log.Printf("Failed to send message: %v", errSend)
				return
			}
			return
		}

		if len(taskList.Tasks) == 0 {
			_, errSend := bot.Send(tgbot.NewMessage(callback.Message.Chat.ID, "У вас нет задач для удаления"))
			if errSend != nil {
				log.Printf("Failed to send message: %v", errSend)
				return
			}
			return
		}

		var taskButtons [][]tgbot.InlineKeyboardButton
		for _, t := range taskList.Tasks {
			taskButtons = append(taskButtons, []tgbot.InlineKeyboardButton{
				tgbot.NewInlineKeyboardButtonData(t.TaskText, "delete_"+strconv.FormatInt(t.TaskId, 10)),
			})
		}

		inlineKeyboard := tgbot.NewInlineKeyboardMarkup(taskButtons...)
		msg := tgbot.NewMessage(callback.Message.Chat.ID, "Выберите задачу для удаления:")
		msg.ReplyMarkup = inlineKeyboard
		_, errSend := bot.Send(msg)
		if errSend != nil {
			log.Printf("Failed to send message: %v", errSend)
			return
		}
		return
	case "cancel_delete":
		inlineKeyboard := MakeButtons()
		msg := tgbot.NewMessage(callback.Message.Chat.ID, "Удаление отменено")
		msg.ReplyMarkup = inlineKeyboard
		_, errSend := bot.Send(msg)
		if errSend != nil {
			log.Printf("Failed to send message: %v", errSend)
			return
		}
		return
	}

	if strings.HasPrefix(callback.Data, "task_") {
		taskID := strings.TrimPrefix(callback.Data, "task_")
		num, err := strconv.ParseInt(taskID, 10, 64)
		if err != nil {
			log.Fatalf("Ошибка конвертации: %v", err)
			return
		}
		userTaskProgressUpdateStates[callback.From.ID] = &TaskProgressUpdateState{
			TaskId: num,
			Step:   1,
		}

		msg := tgbot.NewMessage(callback.Message.Chat.ID, "Введите новый прогресс задачи в % (0-100):")
		_, errSend := bot.Send(msg)
		if errSend != nil {
			log.Printf("Failed to send message: %v", errSend)
			return
		}
	}
	if strings.HasPrefix(callback.Data, "delete_") {
		taskIdStr := strings.TrimPrefix(callback.Data, "delete_")
		taskId, err := strconv.ParseInt(taskIdStr, 10, 64)
		if err != nil {
			log.Printf("Invalid task ID: %v", err)
			_, errSend := bot.Send(tgbot.NewMessage(callback.Message.Chat.ID, "Ошибка: неверный ID задачи"))
			if errSend != nil {
				log.Printf("Failed to send message: %v", errSend)
				return
			}
			return
		}

		inlineKeyboard := tgbot.NewInlineKeyboardMarkup(
			[]tgbot.InlineKeyboardButton{
				tgbot.NewInlineKeyboardButtonData("✅ Да", "confirm_delete_"+strconv.FormatInt(taskId, 10)),
				tgbot.NewInlineKeyboardButtonData("❌ Отмена", "cancel_delete"),
			},
		)

		msg := tgbot.NewMessage(callback.Message.Chat.ID, "Вы уверены, что хотите удалить задачу?")
		msg.ReplyMarkup = inlineKeyboard
		_, errSend := bot.Send(msg)
		if errSend != nil {
			log.Printf("Failed to send message: %v", errSend)
			return
		}
		return
	}
	if strings.HasPrefix(callback.Data, "confirm_delete_") {
		taskIdStr := strings.TrimPrefix(callback.Data, "confirm_delete_")
		taskId, err := strconv.ParseInt(taskIdStr, 10, 64)
		if err != nil {
			log.Printf("Invalid task ID: %v", err)
			_, errSend := bot.Send(tgbot.NewMessage(callback.Message.Chat.ID, "Ошибка: неверный ID задачи"))
			if errSend != nil {
				log.Printf("Failed to send message: %v", errSend)
				return
			}
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		delReq := &task.DeleteTaskRequest{TaskId: taskId}
		_, err = taskService.Client.DeleteTask(ctx, delReq)
		if err != nil {
			log.Printf("Error deleting task: %v", err)
			_, errSend := bot.Send(tgbot.NewMessage(callback.Message.Chat.ID, "Ошибка: не удалось удалить задачу"))
			if errSend != nil {
				log.Printf("Failed to send message: %v", errSend)
				return
			}
			return
		}
		inlineKeyboard := MakeButtons()
		msg := tgbot.NewMessage(callback.Message.Chat.ID, "Задача успешно удалена")
		msg.ReplyMarkup = inlineKeyboard
		_, errSend := bot.Send(msg)
		if errSend != nil {
			log.Printf("Failed to send message: %v", errSend)
			return
		}
		return
	}
}

func getTaskProgressEmoji(progress int32) string {
	switch {
	case progress < 25:
		return "🥵"
	case progress < 50:
		return "🥶"
	case progress < 75:
		return "😶‍🌫"
	case progress < 100:
		return "😉"
	default:
		return "✅"
	}
}
