package catalog

import (
	"sort"
	"strings"
	"unicode/utf8"

	"rasgui/internal/models"

	"golang.org/x/text/encoding/charmap"
)

func Catalog() []models.Operation {
	ops := []models.Operation{
		op("rac.cluster.list", "Список кластеров", "rac", "cluster", []string{"list"}, "Получение списка кластеров", "low", hostOnly()...),
		op("rac.cluster.info", "Информация о кластере", "rac", "cluster", []string{"info"}, "Получение информации о кластере", "low", append(hostOnly(), clusterArg())...),
		op("rac.cluster.insert", "Регистрация кластера", "rac", "cluster", []string{"insert"}, "Регистрация нового кластера", "high", join(hostOnly(), []models.ParamSpec{cliName(common("cluster-host", "Хост кластера", true), "host"), cliName(common("cluster-port", "Порт кластера", true), "port")}, extraArgs())...),
		op("rac.cluster.update", "Обновление кластера", "rac", "cluster", []string{"update"}, "Обновление параметров кластера", "medium", join(clusterAuthParams(), extraArgs())...),
		op("rac.cluster.remove", "Удаление кластера", "rac", "cluster", []string{"remove"}, "Удаление кластера", "critical", join(clusterAuthParams(), extraArgs())...),
		op("rac.cluster.admin.list", "Администраторы кластера", "rac", "cluster", []string{"admin", "list"}, "Список администраторов кластера", "low", clusterAuthParams()...),
		op("rac.cluster.admin.register", "Добавить администратора кластера", "rac", "cluster", []string{"admin", "register"}, "Регистрация администратора кластера", "high", append(clusterAuthParams(), common("name", "Имя администратора", true), password("pwd", "Пароль администратора"), selectArg("auth", "Способ аутентификации", false, "pwd", "pwd,os"), common("os-user", "Пользователь ОС", false), common("agent-user", "Администратор агента", false), password("agent-pwd", "Пароль администратора агента"))...),
		op("rac.cluster.admin.remove", "Удалить администратора кластера", "rac", "cluster", []string{"admin", "remove"}, "Удаление администратора кластера", "high", append(clusterAuthParams(), common("name", "Имя администратора", true))...),
		op("rac.agent.version", "Версия агента", "rac", "agent", []string{"version"}, "Получение версии агента", "low", hostOnly()...),
		op("rac.manager.list", "Список менеджеров", "rac", "manager", []string{"list"}, "Получение списка менеджеров", "low", clusterAuthParams()...),
		op("rac.manager.info", "Информация о менеджере", "rac", "manager", []string{"info"}, "Получение информации о менеджере", "low", append(clusterAuthParams(), common("manager", "UUID менеджера", true))...),
		op("rac.server.list", "Список рабочих серверов", "rac", "server", []string{"list"}, "Получение списка рабочих серверов", "low", clusterAuthParams()...),
		op("rac.server.info", "Информация о рабочем сервере", "rac", "server", []string{"info"}, "Получение информации о рабочем сервере", "low", append(clusterAuthParams(), common("server", "UUID сервера", true))...),
		op("rac.server.insert", "Регистрация рабочего сервера", "rac", "server", []string{"insert"}, "Добавление рабочего сервера", "high", join(clusterAuthParams(), []models.ParamSpec{common("agent-host", "Хост агента", true), common("agent-port", "Порт агента", true), common("port-range", "Диапазон портов", true)}, extraArgs())...),
		op("rac.server.update", "Изменение рабочего сервера", "rac", "server", []string{"update"}, "Изменение параметров рабочего сервера", "medium", join(clusterAuthParams(), []models.ParamSpec{common("server", "UUID сервера", true)}, extraArgs())...),
		op("rac.server.remove", "Удаление рабочего сервера", "rac", "server", []string{"remove"}, "Удаление рабочего сервера", "critical", append(clusterAuthParams(), common("server", "UUID сервера", true))...),
		op("rac.process.list", "Список рабочих процессов", "rac", "process", []string{"list"}, "Получение списка рабочих процессов", "low", clusterAuthParams()...),
		op("rac.process.info", "Информация о рабочем процессе", "rac", "process", []string{"info"}, "Получение информации о рабочем процессе", "low", append(clusterAuthParams(), common("process", "UUID процесса", true))...),
		op("rac.process.turn-off", "Отключение рабочего процесса", "rac", "process", []string{"turn-off"}, "Вывод рабочего процесса из обслуживания", "high", append(clusterAuthParams(), common("process", "UUID процесса", true))...),
		op("rac.service.list", "Список сервисов", "rac", "service", []string{"list"}, "Получение списка сервисов", "low", clusterAuthParams()...),
		op("rac.infobase.summary.list", "Краткий список ИБ", "rac", "infobase", []string{"summary", "list"}, "Получение краткой информации по ИБ", "low", clusterAuthParams()...),
		op("rac.infobase.summary.info", "Краткая информация по ИБ", "rac", "infobase", []string{"summary", "info"}, "Получение краткой информации об ИБ", "low", append(clusterAuthParams(), infobaseSelector()...)...),
		op("rac.infobase.summary.update", "Обновить краткую информацию ИБ", "rac", "infobase", []string{"summary", "update"}, "Обновление краткой информации ИБ", "medium", append(clusterAuthParams(), append(infobaseSelector(), common("descr", "Описание", false))...)...),
		op("rac.infobase.info", "Информация об ИБ", "rac", "infobase", []string{"info"}, "Полная информация об ИБ", "low", append(clusterAuthParams(), append(infobaseSelector(), infobaseCreds()...)...)...),
		op("rac.infobase.create", "Создать ИБ", "rac", "infobase", []string{"create"}, "Создание информационной базы", "critical", join(clusterAuthParams(), []models.ParamSpec{common("name", "Имя ИБ", true), selectArg("dbms", "СУБД", true, "MSSQLServer", "PostgreSQL", "IBMDB2", "OracleDatabase"), common("db-server", "Сервер БД", true), common("db-name", "Имя БД", true), common("locale", "Локаль", true), boolArg("create-database", "Создать базу данных"), common("db-user", "Пользователь БД", false), password("db-pwd", "Пароль БД")}, extraArgs())...),
		op("rac.infobase.update", "Изменить ИБ", "rac", "infobase", []string{"update"}, "Обновление параметров информационной базы", "high", join(clusterAuthParams(), infobaseSelector(), infobaseCreds(), extraArgs())...),
		op("rac.infobase.drop", "Удалить ИБ", "rac", "infobase", []string{"drop"}, "Удаление информационной базы", "critical", append(clusterAuthParams(), append(infobaseSelector(), append(infobaseCreds(), boolArg("drop-database", "Удалить базу данных"), boolArg("clear-database", "Очистить базу данных"))...)...)...),
		op("rac.connection.list", "Список соединений", "rac", "connection", []string{"list"}, "Получение списка соединений", "low", append(clusterAuthParams(), optionalInfobase()...)...),
		op("rac.connection.info", "Информация о соединении", "rac", "connection", []string{"info"}, "Получение информации о соединении", "low", append(clusterAuthParams(), common("connection", "UUID соединения", true))...),
		op("rac.connection.disconnect", "Отключить соединение", "rac", "connection", []string{"disconnect"}, "Принудительное отключение соединения", "high", append(clusterAuthParams(), append([]models.ParamSpec{common("process", "UUID рабочего процесса", true), common("connection", "UUID соединения", true)}, infobaseCreds()...)...)...),
		op("rac.session.list", "Список сеансов", "rac", "session", []string{"list"}, "Получение списка сеансов", "low", append(clusterAuthParams(), optionalInfobase()...)...),
		op("rac.session.info", "Информация о сеансе", "rac", "session", []string{"info"}, "Получение информации о сеансе", "low", append(clusterAuthParams(), common("session", "UUID сеанса", true))...),
		op("rac.session.terminate", "Завершить сеанс", "rac", "session", []string{"terminate"}, "Принудительное завершение сеанса", "high", append(clusterAuthParams(), common("session", "UUID сеанса", true), common("error-message", "Сообщение пользователю", false))...),
		op("rac.session.interrupt-call", "Прервать серверный вызов", "rac", "session", []string{"interrupt-current-server-call"}, "Прерывание текущего серверного вызова", "high", append(clusterAuthParams(), common("session", "UUID сеанса", true))...),
		op("rac.lock.list", "Список блокировок", "rac", "lock", []string{"list"}, "Получение списка блокировок", "low", append(clusterAuthParams(), optionalInfobase()...)...),
		op("rac.rule.list", "Список требований назначения", "rac", "rule", []string{"list"}, "Получение списка требований назначения", "low", append(clusterAuthParams(), common("server", "UUID сервера", true))...),
		op("rac.rule.info", "Информация о требовании назначения", "rac", "rule", []string{"info"}, "Получение информации о требовании назначения", "low", append(clusterAuthParams(), common("server", "UUID сервера", true), common("rule", "UUID требования", true))...),
		op("rac.rule.apply", "Применить требования назначения", "rac", "rule", []string{"apply"}, "Применение требований назначения", "medium", append(clusterAuthParams(), selectArg("mode", "Режим применения", false, "full", "partial"))...),
		op("rac.rule.insert", "Добавить требование назначения", "rac", "rule", []string{"insert"}, "Вставка нового требования назначения", "high", join(clusterAuthParams(), []models.ParamSpec{common("server", "UUID сервера", true), common("position", "Позиция", true)}, extraArgs())...),
		op("rac.rule.update", "Изменить требование назначения", "rac", "rule", []string{"update"}, "Обновление требования назначения", "high", join(clusterAuthParams(), []models.ParamSpec{common("server", "UUID сервера", true), common("rule", "UUID требования", true), common("position", "Позиция", true)}, extraArgs())...),
		op("rac.rule.remove", "Удалить требование назначения", "rac", "rule", []string{"remove"}, "Удаление требования назначения", "critical", append(clusterAuthParams(), common("server", "UUID сервера", true), common("rule", "UUID требования", true))...),
		op("rac.profile.list", "Список профилей безопасности", "rac", "profile", []string{"list"}, "Получение списка профилей безопасности", "low", clusterAuthParams()...),
		op("rac.profile.update", "Создать/обновить профиль безопасности", "rac", "profile", []string{"update"}, "Создание или изменение профиля безопасности", "high", join(clusterAuthParams(), []models.ParamSpec{common("name", "Имя профиля", true)}, extraArgs())...),
		op("rac.profile.remove", "Удалить профиль безопасности", "rac", "profile", []string{"remove"}, "Удаление профиля безопасности", "critical", append(clusterAuthParams(), common("name", "Имя профиля", true))...),
		op("rac.counter.list", "Список счетчиков", "rac", "counter", []string{"list"}, "Получение списка счетчиков", "low", clusterAuthParams()...),
		op("rac.counter.info", "Информация о счетчике", "rac", "counter", []string{"info"}, "Информация по счетчику", "low", append(clusterAuthParams(), common("counter", "Имя счетчика", true))...),
		op("rac.counter.update", "Создать/обновить счетчик", "rac", "counter", []string{"update"}, "Создание или изменение счетчика", "high", join(clusterAuthParams(), []models.ParamSpec{common("name", "Имя счетчика", true)}, extraArgs())...),
		op("rac.counter.values", "Текущие значения счетчика", "rac", "counter", []string{"values"}, "Текущие значения счетчика", "low", append(clusterAuthParams(), common("counter", "Имя счетчика", true), common("object", "Фильтр объекта", false))...),
		op("rac.counter.remove", "Удалить счетчик", "rac", "counter", []string{"remove"}, "Удаление счетчика", "critical", append(clusterAuthParams(), common("name", "Имя счетчика", true))...),
		op("rac.counter.clear", "Очистить значения счетчика", "rac", "counter", []string{"clear"}, "Очистка накопленных значений", "high", append(clusterAuthParams(), common("counter", "Имя счетчика", true), common("object", "Фильтр объекта", false))...),
		op("rac.counter.accumulated-values", "Накопленные значения счетчика", "rac", "counter", []string{"accumulated-values"}, "Получение накопленных значений", "low", append(clusterAuthParams(), common("counter", "Имя счетчика", true), common("object", "Фильтр объекта", false))...),
		op("rac.limit.list", "Список ограничений", "rac", "limit", []string{"list"}, "Получение списка ограничений", "low", clusterAuthParams()...),
		op("rac.limit.info", "Информация об ограничении", "rac", "limit", []string{"info"}, "Получение информации об ограничении", "low", append(clusterAuthParams(), common("limit", "Имя ограничения", true))...),
		op("rac.limit.update", "Создать/обновить ограничение", "rac", "limit", []string{"update"}, "Создание или обновление ограничения", "high", join(clusterAuthParams(), []models.ParamSpec{common("name", "Имя ограничения", true)}, extraArgs())...),
		op("rac.limit.remove", "Удалить ограничение", "rac", "limit", []string{"remove"}, "Удаление ограничения", "critical", append(clusterAuthParams(), common("name", "Имя ограничения", true))...),
		op("rac.service-setting.list", "Список настроек сервисов", "rac", "service-setting", []string{"list"}, "Получение списка настроек сервисов", "low", clusterAuthParams()...),
		op("rac.service-setting.info", "Информация о настройке сервиса", "rac", "service-setting", []string{"info"}, "Получение информации о настройке сервиса", "low", append(clusterAuthParams(), common("service-setting", "UUID настройки", true))...),
		op("rac.service-setting.insert", "Добавить настройку сервиса", "rac", "service-setting", []string{"insert"}, "Добавление настройки сервиса", "high", join(clusterAuthParams(), extraArgs())...),
		op("rac.service-setting.update", "Обновить настройку сервиса", "rac", "service-setting", []string{"update"}, "Изменение настройки сервиса", "high", join(clusterAuthParams(), []models.ParamSpec{common("service-setting", "UUID настройки", true)}, extraArgs())...),
		op("rac.service-setting.get-dirs", "Получить каталоги переноса сервиса", "rac", "service-setting", []string{"get-service-data-dirs-for-transfer"}, "Получение каталогов данных сервиса для переноса", "low", append(clusterAuthParams(), common("service-setting", "UUID настройки", true))...),
		op("rac.service-setting.remove", "Удалить настройку сервиса", "rac", "service-setting", []string{"remove"}, "Удаление настройки сервиса", "critical", append(clusterAuthParams(), common("service-setting", "UUID настройки", true))...),
		op("rac.service-setting.apply", "Применить настройку сервиса", "rac", "service-setting", []string{"apply"}, "Применение настройки сервиса", "high", append(clusterAuthParams(), common("service-setting", "UUID настройки", true))...),
		op("rac.binary-data-storage.list", "Список хранилищ двоичных данных", "rac", "binary-data-storage", []string{"list"}, "Получение списка хранилищ двоичных данных", "low", clusterAuthParams()...),
		op("rac.binary-data-storage.info", "Информация о хранилище двоичных данных", "rac", "binary-data-storage", []string{"info"}, "Получение информации о хранилище", "low", append(clusterAuthParams(), common("binary-data-storage", "UUID хранилища", true))...),
		op("rac.binary-data-storage.create-full-backup", "Полный backup BDS", "rac", "binary-data-storage", []string{"create-full-backup"}, "Создание полного резервного копирования", "high", join(clusterAuthParams(), []models.ParamSpec{common("binary-data-storage", "UUID хранилища", true)}, extraArgs())...),
		op("rac.binary-data-storage.create-diff-backup", "Дифференциальный backup BDS", "rac", "binary-data-storage", []string{"create-diff-backup"}, "Создание дифференциального резервного копирования", "high", join(clusterAuthParams(), []models.ParamSpec{common("binary-data-storage", "UUID хранилища", true)}, extraArgs())...),
		op("rac.binary-data-storage.load-full-backup", "Восстановить полный backup BDS", "rac", "binary-data-storage", []string{"load-full-backup"}, "Загрузка полного резервного копирования", "critical", join(clusterAuthParams(), []models.ParamSpec{common("binary-data-storage", "UUID хранилища", true)}, extraArgs())...),
		op("rac.binary-data-storage.load-diff-backup", "Восстановить diff backup BDS", "rac", "binary-data-storage", []string{"load-diff-backup"}, "Загрузка дифференциального резервного копирования", "critical", join(clusterAuthParams(), []models.ParamSpec{common("binary-data-storage", "UUID хранилища", true)}, extraArgs())...),
		op("rac.binary-data-storage.clear-unused-space", "Очистить неиспользуемое место BDS", "rac", "binary-data-storage", []string{"clear-unused-space"}, "Очистка неиспользуемого места", "high", append(clusterAuthParams(), common("binary-data-storage", "UUID хранилища", true))...),
		op("ras.cluster.run", "Запуск `ras cluster`", "ras", "cluster", nil, "Запуск сервера администрирования в режиме cluster", "critical", common("target_host", "Адрес агента кластера", true), common("port", "Порт RAS", false), common("monitor-address", "Адрес monitor", false), common("monitor-port", "Порт monitor", false), common("monitor-base", "Базовый путь monitor", false), boolArg("service", "Режим сервиса Windows")),
	}

	for i := range ops {
		ops[i].Title = repairEncoding(ops[i].Title)
		ops[i].Description = repairEncoding(ops[i].Description)
		for j := range ops[i].Params {
			ops[i].Params[j].Label = repairEncoding(ops[i].Params[j].Label)
			ops[i].Params[j].Description = repairEncoding(ops[i].Params[j].Description)
		}
	}

	sort.Slice(ops, func(i, j int) bool { return ops[i].ID < ops[j].ID })
	return ops
}

func Find(id string) (models.Operation, bool) {
	for _, item := range Catalog() {
		if item.ID == id {
			return item, true
		}
	}
	return models.Operation{}, false
}

func op(id, title, utility, mode string, subcommands []string, desc, risk string, params ...models.ParamSpec) models.Operation {
	return models.Operation{
		ID:          id,
		Title:       repairEncoding(title),
		Utility:     utility,
		Mode:        mode,
		Subcommands: subcommands,
		Description: repairEncoding(desc),
		RiskLevel:   risk,
		Params:      params,
	}
}

func hostOnly() []models.ParamSpec {
	return []models.ParamSpec{common("host", "Адрес сервера администрирования", false), common("admin_port", "Порт сервера администрирования", false)}
}

func clusterAuthParams() []models.ParamSpec {
	return []models.ParamSpec{
		common("host", "Адрес сервера администрирования", false),
		common("admin_port", "Порт сервера администрирования", false),
		clusterArg(),
		common("cluster-user", "Администратор кластера", false),
		password("cluster-pwd", "Пароль администратора кластера"),
	}
}

func clusterArg() models.ParamSpec { return common("cluster", "UUID кластера", true) }

func infobaseSelector() []models.ParamSpec {
	return []models.ParamSpec{common("infobase", "UUID ИБ", false), common("name", "Имя ИБ", false)}
}

func infobaseCreds() []models.ParamSpec {
	return []models.ParamSpec{common("infobase-user", "Администратор ИБ", false), password("infobase-pwd", "Пароль администратора ИБ")}
}

func optionalInfobase() []models.ParamSpec {
	return []models.ParamSpec{common("infobase", "UUID ИБ", false), common("name", "Имя ИБ", false)}
}

func common(name, label string, required bool) models.ParamSpec {
	return models.ParamSpec{Name: name, Label: repairEncoding(label), Type: models.ParamString, Required: required}
}

func cliName(spec models.ParamSpec, argName string) models.ParamSpec {
	spec.ArgName = argName
	return spec
}

func password(name, label string) models.ParamSpec {
	return models.ParamSpec{Name: name, Label: repairEncoding(label), Type: models.ParamPassword}
}

func boolArg(name, label string) models.ParamSpec {
	return models.ParamSpec{Name: name, Label: repairEncoding(label), Type: models.ParamBool}
}

func selectArg(name, label string, required bool, options ...string) models.ParamSpec {
	return models.ParamSpec{Name: name, Label: repairEncoding(label), Type: models.ParamSelect, Required: required, Options: options}
}

func extraArgs() []models.ParamSpec {
	return []models.ParamSpec{{Name: "extra_args", Label: "Дополнительные аргументы", Type: models.ParamString, Description: "Дополнительные параметры в формате командной строки"}}
}

func join(groups ...[]models.ParamSpec) []models.ParamSpec {
	var result []models.ParamSpec
	for _, group := range groups {
		result = append(result, group...)
	}
	return result
}

func repairEncoding(value string) string {
	if value == "" {
		return value
	}
	if !strings.ContainsAny(value, "РСЃЌЏ") {
		return value
	}
	encoded, err := charmap.Windows1251.NewEncoder().Bytes([]byte(value))
	if err != nil || !utf8.Valid(encoded) {
		return value
	}
	return string(encoded)
}
