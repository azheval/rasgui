(function () {
  const operations = window.rasguiOperations || [];
  const profiles = window.rasguiProfiles || [];
  const favorites = window.rasguiFavorites || [];
  const formValues = window.rasguiFormValues || {};
  const defaults = window.rasguiDefaults || {};
  const locale = window.rasguiLocale || {};
  const lang = window.rasguiLang || "ru";

  const select = document.getElementById("operation_id");
  const primaryFields = document.getElementById("primary_fields");
  const advancedFields = document.getElementById("advanced_fields");
  const advancedFieldsBox = document.getElementById("advanced_fields_box");
  const entityFilters = document.getElementById("entity_filters");
  const quickActions = document.getElementById("quick_actions");
  const guidedScenarios = document.getElementById("guided_scenarios");
  const operatorFlowHost = document.getElementById("operator_flow_host");
  const wizardAssistant = document.getElementById("wizard_assistant");
  const operationDetails = document.getElementById("operation_details");
  const resultContextHost = document.getElementById("result_context_host");
  const journeySummary = document.getElementById("journey_summary");
  const humanActionList = document.getElementById("human_action_list");
  const friendlyActionLabel = document.getElementById("friendly_action_label");
  const friendlyActionHint = document.getElementById("friendly_action_hint");
  const playbookLabel = document.getElementById("playbook_label");
  const playbookHint = document.getElementById("playbook_hint");
  const playbookList = document.getElementById("playbook_list");
  const fieldGuideLabel = document.getElementById("field_guide_label");
  const fieldGuideHint = document.getElementById("field_guide_hint");
  const fieldGuide = document.getElementById("field_guide");
  const smartWizardHost = document.getElementById("smart_wizard_host");
  const wizardArgsJsonInput = document.getElementById("wizard_args_json");
  const technicalCommandLabel = document.getElementById("technical_command_label");
  const technicalCommandHint = document.getElementById("technical_command_hint");
  const execForm = document.getElementById("exec-form");
  const profileSelect = document.getElementById("connection_profile");
  const connectionProfileIdInput = document.getElementById("connection_profile_id");
  const favoriteSelect = document.getElementById("favorite_select");
  const favoriteConnection = document.getElementById("favorite_connection_id");
  const favoriteIdInput = document.getElementById("favorite_id");
  const favoriteNameInput = document.getElementById("favorite_name");
  const favoriteNote = document.getElementById("favorite_note");
  const deleteFavoriteButton = document.getElementById("delete_favorite");
  const catalogSearch = document.getElementById("catalog_search");
  const catalogRows = document.querySelectorAll("#catalog_table tbody tr");
  const catalogEntityFilters = document.getElementById("catalog_entity_filters");

  let activeEntity = "";
  let activePlaybook = "";
  let activeSmartWizardPreset = "";
  let activeFlowId = "";
  let activeFlowStepIndex = 0;

  const scenarioDefinitions = {
    cluster: ["rac.cluster.list", "rac.cluster.info", "rac.cluster.admin.list", "rac.cluster.update"],
    infobase: ["rac.infobase.summary.list", "rac.infobase.info", "rac.infobase.create", "rac.infobase.drop"],
    session: ["rac.session.list", "rac.session.info", "rac.session.terminate", "rac.session.interrupt-call"],
    connection: ["rac.connection.list", "rac.connection.info", "rac.connection.disconnect"],
    server: ["rac.server.list", "rac.server.info", "rac.server.update"],
  };

  const playbookDefinitions = {
    cluster: [
      { id: "inspect", opIds: ["rac.cluster.list", "rac.cluster.info"], labelKey: "playbookInspect", hintKey: "playbookInspectHint" },
      { id: "security", opIds: ["rac.cluster.admin.list", "rac.cluster.admin.register", "rac.cluster.admin.remove"], labelKey: "playbookSecurity", hintKey: "playbookSecurityHint" },
      { id: "change", opIds: ["rac.cluster.update"], labelKey: "playbookChange", hintKey: "playbookChangeHint" },
      { id: "lifecycle", opIds: ["rac.cluster.insert", "rac.cluster.remove"], labelKey: "playbookLifecycle", hintKey: "playbookLifecycleHint" },
    ],
    infobase: [
      { id: "inspect", opIds: ["rac.infobase.summary.list", "rac.infobase.summary.info", "rac.infobase.info"], labelKey: "playbookInspect", hintKey: "playbookInspectHint" },
      { id: "change", opIds: ["rac.infobase.summary.update", "rac.infobase.update"], labelKey: "playbookChange", hintKey: "playbookChangeHint" },
      { id: "lifecycle", opIds: ["rac.infobase.create", "rac.infobase.drop"], labelKey: "playbookLifecycle", hintKey: "playbookLifecycleHint" },
    ],
    session: [
      { id: "inspect", opIds: ["rac.session.list", "rac.session.info"], labelKey: "playbookInspect", hintKey: "playbookInspectHint" },
      { id: "control", opIds: ["rac.session.terminate", "rac.session.interrupt-call"], labelKey: "playbookControl", hintKey: "playbookControlHint" },
    ],
    connection: [
      { id: "inspect", opIds: ["rac.connection.list", "rac.connection.info"], labelKey: "playbookInspect", hintKey: "playbookInspectHint" },
      { id: "control", opIds: ["rac.connection.disconnect"], labelKey: "playbookControl", hintKey: "playbookControlHint" },
    ],
    server: [
      { id: "inspect", opIds: ["rac.server.list", "rac.server.info"], labelKey: "playbookInspect", hintKey: "playbookInspectHint" },
      { id: "change", opIds: ["rac.server.update"], labelKey: "playbookChange", hintKey: "playbookChangeHint" },
      { id: "lifecycle", opIds: ["rac.server.insert", "rac.server.remove"], labelKey: "playbookLifecycle", hintKey: "playbookLifecycleHint" },
    ],
  };

  const operatorFlowDefinitions = [
    {
      id: "infobase_maintenance",
      entity: "infobase",
      title: byLang({ ru: "Обслуживание ИБ", be: "Абслугоўванне ІБ", en: "Infobase maintenance" }),
      hint: byLang({ ru: "Заблокировать ИБ, проверить сеансы и отключить пользователей.", be: "Заблакіраваць ІБ, праверыць сеансы і адключыць карыстальнікаў.", en: "Lock the infobase, inspect sessions, and disconnect users." }),
      steps: [
        { operationId: "rac.infobase.update", label: byLang({ ru: "Включить блокировку ИБ", be: "Уключыць блакіроўку ІБ", en: "Enable infobase lock" }), required: true },
        { operationId: "rac.session.list", label: byLang({ ru: "Проверить активные сеансы", be: "Праверыць актыўныя сеансы", en: "Inspect active sessions" }), required: false },
        { operationId: "rac.session.terminate", label: byLang({ ru: "Завершить мешающие сеансы", be: "Завяршыць сеансы, што перашкаджаюць", en: "Terminate blocking sessions" }), required: false },
        { operationId: "rac.connection.list", label: byLang({ ru: "Проверить соединения", be: "Праверыць злучэнні", en: "Inspect connections" }), required: false },
        { operationId: "rac.connection.disconnect", label: byLang({ ru: "Отключить оставшиеся соединения", be: "Адключыць астатнія злучэнні", en: "Disconnect remaining connections" }), required: false },
      ],
    },
    {
      id: "infobase_unlock",
      entity: "infobase",
      title: byLang({ ru: "Вывод ИБ из обслуживания", be: "Вывад ІБ з абслугоўвання", en: "Return infobase to service" }),
      hint: byLang({ ru: "Снять блокировку и вернуть рабочий режим.", be: "Зняць блакіроўку і вярнуць працоўны рэжым.", en: "Remove the lock and return the infobase to normal work." }),
      steps: [
        { operationId: "rac.infobase.update", label: byLang({ ru: "Снять блокировку и включить задания", be: "Зняць блакіроўку і ўключыць заданні", en: "Remove lock and enable jobs" }), required: true },
      ],
    },
    {
      id: "session_cleanup",
      entity: "session",
      title: byLang({ ru: "Очистка пользовательских сеансов", be: "Ачыстка карыстальніцкіх сеансаў", en: "Session cleanup" }),
      hint: byLang({ ru: "Найти сеансы и точечно завершить нужные.", be: "Знайсці сеансы і кропкава завяршыць патрэбныя.", en: "Find sessions and terminate only the required ones." }),
      steps: [
        { operationId: "rac.session.list", label: byLang({ ru: "Получить список сеансов", be: "Атрымаць спіс сеансаў", en: "List sessions" }), required: true },
        { operationId: "rac.session.info", label: byLang({ ru: "Проверить детали сеанса", be: "Праверыць дэталі сеансу", en: "Inspect session details" }), required: false },
        { operationId: "rac.session.terminate", label: byLang({ ru: "Завершить выбранный сеанс", be: "Завяршыць выбраны сеанс", en: "Terminate selected session" }), required: true },
      ],
    },
    {
      id: "connection_cleanup",
      entity: "connection",
      title: byLang({ ru: "Очистка соединений", be: "Ачыстка злучэнняў", en: "Connection cleanup" }),
      hint: byLang({ ru: "Посмотреть соединения и разорвать проблемные.", be: "Паглядзець злучэнні і разарваць праблемныя.", en: "Inspect connections and disconnect problematic ones." }),
      steps: [
        { operationId: "rac.connection.list", label: byLang({ ru: "Получить список соединений", be: "Атрымаць спіс злучэнняў", en: "List connections" }), required: true },
        { operationId: "rac.connection.info", label: byLang({ ru: "Проверить детали соединения", be: "Праверыць дэталі злучэння", en: "Inspect connection details" }), required: false },
        { operationId: "rac.connection.disconnect", label: byLang({ ru: "Отключить выбранное соединение", be: "Адключыць выбранае злучэнне", en: "Disconnect selected connection" }), required: true },
      ],
    },
  ];

  function byLang(values) {
    if (!values) return "";
    return values[lang] || values.ru || values.en || "";
  }

  const smartWizardDefinitions = {
    "rac.infobase.update": {
      builder: "infobase_lock",
      title: byLang({
        ru: "Мастер блокировки и режима обслуживания",
        be: "Майстар блакіроўкі і рэжыму абслугоўвання",
        en: "Maintenance and lock wizard",
      }),
      hint: byLang({
        ru: "Настройте блокировку ИБ в понятной форме, а нужные параметры rac будут собраны автоматически.",
        be: "Наладзьце блакіроўку ІБ у зразумелай форме, а патрэбныя параметры rac будуць сабраныя аўтаматычна.",
        en: "Configure infobase lock settings in a guided form and the required rac parameters will be generated automatically.",
      }),
      previewTitle: byLang({
        ru: "Что будет добавлено к команде",
        be: "Што будзе дададзена да каманды",
        en: "Arguments that will be added",
      }),
      presetsTitle: byLang({
        ru: "Готовые сценарии",
        be: "Гатовыя сцэнарыі",
        en: "Ready scenarios",
      }),
      fieldsTitle: byLang({
        ru: "Параметры блокировки",
        be: "Параметры блакіроўкі",
        en: "Lock parameters",
      }),
      presets: [
        {
          id: "maintenance_lock",
          title: byLang({ ru: "Заблокировать ИБ для обслуживания", be: "Заблакіраваць ІБ для абслугоўвання", en: "Lock for maintenance" }),
          hint: byLang({ ru: "Запретить вход и остановить регламентные задания", be: "Забараніць уваход і спыніць рэгламентныя заданні", en: "Block logins and scheduled jobs" }),
          values: { wizard_sessions_deny: "true", wizard_scheduled_jobs_deny: "true" },
        },
        {
          id: "sessions_only",
          title: byLang({ ru: "Запретить вход в ИБ", be: "Забараніць уваход у ІБ", en: "Block logins only" }),
          hint: byLang({ ru: "Блокировка начала новых сеансов без остановки заданий", be: "Блакіроўка новых сеансаў без спынення заданняў", en: "Block new sessions without pausing scheduled jobs" }),
          values: { wizard_sessions_deny: "true", wizard_scheduled_jobs_deny: "false" },
        },
        {
          id: "jobs_only",
          title: byLang({ ru: "Остановить регламентные задания", be: "Спыніць рэгламентныя заданні", en: "Pause scheduled jobs" }),
          hint: byLang({ ru: "Оставить вход разрешенным, но отключить задания", be: "Пакінуць уваход дазволеным, але адключыць заданні", en: "Keep logins allowed but stop scheduled jobs" }),
          values: { wizard_sessions_deny: "false", wizard_scheduled_jobs_deny: "true" },
        },
        {
          id: "unlock_all",
          title: byLang({ ru: "Снять ограничения", be: "Зняць абмежаванні", en: "Remove restrictions" }),
          hint: byLang({ ru: "Разрешить вход и выполнение регламентных заданий", be: "Дазволіць уваход і выкананне рэгламентных заданняў", en: "Allow logins and scheduled jobs again" }),
          values: { wizard_sessions_deny: "false", wizard_scheduled_jobs_deny: "false", wizard_denied_message: "", wizard_permission_code: "", wizard_denied_parameter: "", wizard_denied_from: "", wizard_denied_to: "" },
        },
      ],
      fields: [
        { name: "wizard_sessions_deny", type: "bool", label: byLang({ ru: "Запретить вход в ИБ", be: "Забараніць уваход у ІБ", en: "Block new sessions" }) },
        { name: "wizard_scheduled_jobs_deny", type: "bool", label: byLang({ ru: "Запретить регламентные задания", be: "Забараніць рэгламентныя заданні", en: "Block scheduled jobs" }) },
        { name: "wizard_denied_message", type: "text", label: byLang({ ru: "Сообщение пользователю", be: "Паведамленне карыстальніку", en: "User message" }), placeholder: byLang({ ru: "Например: База закрыта на обслуживание", be: "Напрыклад: База закрыта на абслугоўванне", en: "For example: Infobase is under maintenance" }) },
        { name: "wizard_permission_code", type: "text", label: byLang({ ru: "Код разрешения доступа", be: "Код дазволу доступу", en: "Permission code" }), placeholder: byLang({ ru: "Код для входа при блокировке", be: "Код для ўваходу пры блакіроўцы", en: "Code that bypasses the lock" }) },
        { name: "wizard_denied_parameter", type: "text", label: byLang({ ru: "Параметр блокировки", be: "Параметр блакіроўкі", en: "Lock parameter" }), placeholder: byLang({ ru: "Необязательно", be: "Неабавязкова", en: "Optional" }) },
        { name: "wizard_denied_from", type: "datetime-local", label: byLang({ ru: "Блокировка действует с", be: "Блакіроўка дзейнічае з", en: "Lock starts at" }) },
        { name: "wizard_denied_to", type: "datetime-local", label: byLang({ ru: "Блокировка действует до", be: "Блакіроўка дзейнічае да", en: "Lock ends at" }) },
      ],
    },
    "rac.infobase.create": {
      title: byLang({
        ru: "Мастер создания ИБ",
        be: "Майстар стварэння ІБ",
        en: "Infobase creation wizard",
      }),
      hint: byLang({
        ru: "Соберите новую ИБ по сценарию, не переключаясь между десятком полей.",
        be: "Збярыце новую ІБ па сцэнарыі, не пераключаючыся паміж дзясяткам палёў.",
        en: "Create a new infobase through a scenario instead of filling a long raw form.",
      }),
      presetsTitle: byLang({ ru: "Типовой сценарий", be: "Тыпавы сцэнар", en: "Common scenario" }),
      fieldsTitle: byLang({ ru: "Основные параметры ИБ", be: "Асноўныя параметры ІБ", en: "Core infobase settings" }),
      hideParams: ["name", "dbms", "db-server", "db-name", "locale", "create-database", "db-user", "db-pwd", "descr", "scheduled-jobs-deny", "license-distribution"],
      presets: [
        {
          id: "mssql_new",
          title: byLang({ ru: "Новая ИБ в MS SQL", be: "Новая ІБ у MS SQL", en: "New MS SQL infobase" }),
          hint: byLang({ ru: "Создать новую БД и зарегистрировать ИБ", be: "Стварыць новую БД і зарэгістраваць ІБ", en: "Create a new database and register the infobase" }),
          values: { dbms: "MSSQLServer", "create-database": "true", locale: "ru" },
        },
        {
          id: "postgres_new",
          title: byLang({ ru: "Новая ИБ в PostgreSQL", be: "Новая ІБ у PostgreSQL", en: "New PostgreSQL infobase" }),
          hint: byLang({ ru: "Создать новую PostgreSQL БД", be: "Стварыць новую PostgreSQL БД", en: "Create a new PostgreSQL database" }),
          values: { dbms: "PostgreSQL", "create-database": "true", locale: "ru" },
        },
        {
          id: "attach_existing",
          title: byLang({ ru: "Подключить существующую БД", be: "Падключыць існуючую БД", en: "Attach existing database" }),
          hint: byLang({ ru: "Не создавать БД, только зарегистрировать ИБ", be: "Не ствараць БД, толькі зарэгістраваць ІБ", en: "Do not create a DB, only register the infobase" }),
          values: { "create-database": "false" },
        },
      ],
      fields: [
        { name: "name", type: "text", label: byLang({ ru: "Имя ИБ", be: "Імя ІБ", en: "Infobase name" }) },
        { name: "dbms", type: "select", label: byLang({ ru: "СУБД", be: "СУБД", en: "DBMS" }), options: ["MSSQLServer", "PostgreSQL", "IBMDB2", "OracleDatabase"] },
        { name: "db-server", type: "text", label: byLang({ ru: "Сервер БД", be: "Сервер БД", en: "DB server" }) },
        { name: "db-name", type: "text", label: byLang({ ru: "Имя БД", be: "Імя БД", en: "DB name" }) },
        { name: "locale", type: "text", label: byLang({ ru: "Локаль", be: "Лакаль", en: "Locale" }), placeholder: "ru" },
        { name: "create-database", type: "bool", label: byLang({ ru: "Создать БД автоматически", be: "Стварыць БД аўтаматычна", en: "Create database automatically" }) },
        { name: "db-user", type: "text", label: byLang({ ru: "Пользователь БД", be: "Карыстальнік БД", en: "DB user" }) },
        { name: "db-pwd", type: "password", label: byLang({ ru: "Пароль БД", be: "Пароль БД", en: "DB password" }) },
        { name: "descr", type: "text", label: byLang({ ru: "Описание", be: "Апісанне", en: "Description" }) },
      ],
    },
    "rac.infobase.drop": {
      title: byLang({
        ru: "Мастер удаления ИБ",
        be: "Майстар выдалення ІБ",
        en: "Infobase removal wizard",
      }),
      hint: byLang({
        ru: "Выберите сценарий удаления, чтобы не вспоминать флаги очистки и удаления БД.",
        be: "Абярыце сцэнар выдалення, каб не ўспамінаць сцяжкі ачысткі і выдалення БД.",
        en: "Choose a removal scenario instead of remembering database cleanup flags.",
      }),
      presetsTitle: byLang({ ru: "Сценарий удаления", be: "Сцэнар выдалення", en: "Removal scenario" }),
      fieldsTitle: byLang({ ru: "Что удаляем", be: "Што выдаляем", en: "What to remove" }),
      hideParams: ["infobase", "name", "infobase-user", "infobase-pwd", "drop-database", "clear-database"],
      presets: [
        {
          id: "unregister_only",
          title: byLang({ ru: "Только снять регистрацию", be: "Толькі зняць рэгістрацыю", en: "Unregister only" }),
          hint: byLang({ ru: "Оставить физическую БД нетронутой", be: "Пакінуць фізічную БД некранутай", en: "Keep the physical database intact" }),
          values: { "drop-database": "false", "clear-database": "false" },
        },
        {
          id: "drop_with_db",
          title: byLang({ ru: "Удалить ИБ и БД", be: "Выдаліць ІБ і БД", en: "Delete infobase and DB" }),
          hint: byLang({ ru: "Полное удаление с физической БД", be: "Поўнае выдаленне з фізічнай БД", en: "Full removal including the database" }),
          values: { "drop-database": "true", "clear-database": "false" },
        },
        {
          id: "clear_db_only",
          title: byLang({ ru: "Очистить БД", be: "Ачысціць БД", en: "Clear database" }),
          hint: byLang({ ru: "Оставить регистрацию, но очистить данные БД", be: "Пакінуць рэгістрацыю, але ачысціць даныя БД", en: "Keep registration but clear the database contents" }),
          values: { "drop-database": "false", "clear-database": "true" },
        },
      ],
      fields: [
        { name: "infobase", type: "text", label: byLang({ ru: "UUID ИБ", be: "UUID ІБ", en: "Infobase UUID" }) },
        { name: "name", type: "text", label: byLang({ ru: "Имя ИБ", be: "Імя ІБ", en: "Infobase name" }) },
        { name: "infobase-user", type: "text", label: byLang({ ru: "Администратор ИБ", be: "Адміністратар ІБ", en: "Infobase admin" }) },
        { name: "infobase-pwd", type: "password", label: byLang({ ru: "Пароль администратора ИБ", be: "Пароль адміністратара ІБ", en: "Infobase admin password" }) },
        { name: "drop-database", type: "bool", label: byLang({ ru: "Удалить физическую БД", be: "Выдаліць фізічную БД", en: "Drop physical database" }) },
        { name: "clear-database", type: "bool", label: byLang({ ru: "Очистить физическую БД", be: "Ачысціць фізічную БД", en: "Clear physical database" }) },
      ],
    },
    "rac.cluster.admin.register": {
      title: byLang({
        ru: "Мастер добавления администратора кластера",
        be: "Майстар дадання адміністратара кластара",
        en: "Cluster admin registration wizard",
      }),
      hint: byLang({
        ru: "Упростите добавление администратора: локальная учетная запись или привязка к ОС.",
        be: "Спрасціце даданне адміністратара: лакальны ўліковы запіс або прывязка да АС.",
        en: "Register a cluster administrator through a clear local or OS-linked flow.",
      }),
      presetsTitle: byLang({ ru: "Сценарий аутентификации", be: "Сцэнар аўтэнтыфікацыі", en: "Authentication scenario" }),
      fieldsTitle: byLang({ ru: "Параметры администратора", be: "Параметры адміністратара", en: "Administrator settings" }),
      hideParams: ["name", "pwd", "auth", "os-user", "agent-user", "agent-pwd"],
      presets: [
        {
          id: "local_auth",
          title: byLang({ ru: "Локальный администратор", be: "Лакальны адміністратар", en: "Local administrator" }),
          hint: byLang({ ru: "Имя и пароль внутри кластера", be: "Імя і пароль унутры кластара", en: "Name and password managed by the cluster" }),
          values: { auth: "pwd" },
        },
        {
          id: "os_auth",
          title: byLang({ ru: "Администратор ОС", be: "Адміністратар АС", en: "OS-backed administrator" }),
          hint: byLang({ ru: "Аутентификация через пользователя ОС", be: "Аўтэнтыфікацыя праз карыстальніка АС", en: "Authenticate through an OS account" }),
          values: { auth: "pwd,os" },
        },
      ],
      fields: [
        { name: "name", type: "text", label: byLang({ ru: "Имя администратора", be: "Імя адміністратара", en: "Administrator name" }) },
        { name: "pwd", type: "password", label: byLang({ ru: "Пароль администратора", be: "Пароль адміністратара", en: "Administrator password" }) },
        { name: "auth", type: "select", label: byLang({ ru: "Способ аутентификации", be: "Спосаб аўтэнтыфікацыі", en: "Authentication mode" }), options: ["pwd", "pwd,os"] },
        { name: "os-user", type: "text", label: byLang({ ru: "Пользователь ОС", be: "Карыстальнік АС", en: "OS user" }) },
        { name: "agent-user", type: "text", label: byLang({ ru: "Администратор агента", be: "Адміністратар агента", en: "Agent administrator" }) },
        { name: "agent-pwd", type: "password", label: byLang({ ru: "Пароль администратора агента", be: "Пароль адміністратара агента", en: "Agent administrator password" }) },
      ],
    },
    "rac.session.terminate": {
      title: byLang({
        ru: "Мастер завершения сеанса",
        be: "Майстар завяршэння сеансу",
        en: "Session termination wizard",
      }),
      hint: byLang({
        ru: "Завершите сеанс через понятный сценарий и при необходимости покажите пользователю причину.",
        be: "Завяршыце сеанс праз зразумелы сцэнар і пры патрэбе пакажыце карыстальніку прычыну.",
        en: "Terminate a session through a guided flow and optionally show the reason to the user.",
      }),
      presetsTitle: byLang({ ru: "Сценарий завершения", be: "Сцэнар завяршэння", en: "Termination scenario" }),
      fieldsTitle: byLang({ ru: "Параметры завершения", be: "Параметры завяршэння", en: "Termination details" }),
      hideParams: ["session", "error-message"],
      presets: [
        {
          id: "silent",
          title: byLang({ ru: "Тихо завершить", be: "Ціха завяршыць", en: "Terminate silently" }),
          hint: byLang({ ru: "Без сообщения пользователю", be: "Без паведамлення карыстальніку", en: "No message shown to the user" }),
          values: { "error-message": "" },
        },
        {
          id: "maintenance",
          title: byLang({ ru: "Завершить на обслуживание", be: "Завяршыць на абслугоўванне", en: "Terminate for maintenance" }),
          hint: byLang({ ru: "Показать понятное сообщение о техработах", be: "Паказаць зразумелае паведамленне пра тэхработы", en: "Show a maintenance notice" }),
          values: { "error-message": "Maintenance window" },
        },
      ],
      fields: [
        { name: "session", type: "text", label: byLang({ ru: "UUID сеанса", be: "UUID сеансу", en: "Session UUID" }) },
        { name: "error-message", type: "text", label: byLang({ ru: "Сообщение пользователю", be: "Паведамленне карыстальніку", en: "User message" }) },
      ],
    },
    "rac.connection.disconnect": {
      title: byLang({
        ru: "Мастер отключения соединения",
        be: "Майстар адключэння злучэння",
        en: "Connection disconnect wizard",
      }),
      hint: byLang({
        ru: "Укажите соединение и рабочий процесс без ручного поиска обязательных параметров.",
        be: "Пакажыце злучэнне і працэс без ручнога пошуку абавязковых параметраў.",
        en: "Pick the connection and worker process without manually collecting required arguments.",
      }),
      presetsTitle: byLang({ ru: "Сценарий отключения", be: "Сцэнар адключэння", en: "Disconnect scenario" }),
      fieldsTitle: byLang({ ru: "Что отключаем", be: "Што адключаем", en: "Connection details" }),
      hideParams: ["process", "connection", "infobase-user", "infobase-pwd"],
      presets: [
        {
          id: "plain_disconnect",
          title: byLang({ ru: "Принудительно отключить", be: "Прымусова адключыць", en: "Force disconnect" }),
          hint: byLang({ ru: "Разорвать соединение сразу", be: "Разарваць злучэнне адразу", en: "Break the connection immediately" }),
          values: {},
        },
        {
          id: "with_ib_admin",
          title: byLang({ ru: "Отключить с правами администратора ИБ", be: "Адключыць з правамі адміністратара ІБ", en: "Disconnect with infobase admin" }),
          hint: byLang({ ru: "Когда сервер требует учетные данные администратора ИБ", be: "Калі сервер патрабуе ўліковыя даныя адміністратара ІБ", en: "Use infobase admin credentials when required" }),
          values: {},
        },
      ],
      fields: [
        { name: "process", type: "text", label: byLang({ ru: "UUID рабочего процесса", be: "UUID рабочага працэсу", en: "Worker process UUID" }) },
        { name: "connection", type: "text", label: byLang({ ru: "UUID соединения", be: "UUID злучэння", en: "Connection UUID" }) },
        { name: "infobase-user", type: "text", label: byLang({ ru: "Администратор ИБ", be: "Адміністратар ІБ", en: "Infobase admin" }) },
        { name: "infobase-pwd", type: "password", label: byLang({ ru: "Пароль администратора ИБ", be: "Пароль адміністратара ІБ", en: "Infobase admin password" }) },
      ],
    },
    "rac.cluster.update": {
      builder: "cluster_update",
      title: byLang({
        ru: "Мастер настройки кластера",
        be: "Майстар наладкі кластара",
        en: "Cluster tuning wizard",
      }),
      hint: byLang({
        ru: "Меняйте ключевые параметры кластера как настройки, а не как набор флагов командной строки.",
        be: "Змяняйце ключавыя параметры кластара як налады, а не як набор сцяжкоў каманднага радка.",
        en: "Change cluster settings as named options instead of a raw flag list.",
      }),
      previewTitle: byLang({ ru: "Параметры обновления", be: "Параметры абнаўлення", en: "Update arguments" }),
      presetsTitle: byLang({ ru: "Профиль настройки", be: "Профіль наладкі", en: "Tuning profile" }),
      fieldsTitle: byLang({ ru: "Параметры кластера", be: "Параметры кластара", en: "Cluster settings" }),
      hideParams: ["extra_args"],
      presets: [
        {
          id: "balanced",
          title: byLang({ ru: "Сбалансированный", be: "Збалансаваны", en: "Balanced" }),
          hint: byLang({ ru: "Производительность с безопасными настройками по умолчанию", be: "Прадукцыйнасць з бяспечнымі наладамі па змоўчанні", en: "Performance with safe defaults" }),
          values: { "wizard_cluster_load_balancing_mode": "performance", "wizard_cluster_kill_problem_processes": "true", "wizard_cluster_allow_access_audit": "false" },
        },
        {
          id: "memory_safe",
          title: byLang({ ru: "Приоритет памяти", be: "Прыярытэт памяці", en: "Memory-priority" }),
          hint: byLang({ ru: "Бережнее к памяти и проблемным процессам", be: "Акуратней з памяццю і праблемнымі працэсамі", en: "More conservative with memory and problematic processes" }),
          values: { "wizard_cluster_load_balancing_mode": "memory", "wizard_cluster_kill_problem_processes": "true" },
        },
      ],
      fields: [
        { name: "wizard_cluster_name", type: "text", label: byLang({ ru: "Имя кластера", be: "Імя кластара", en: "Cluster display name" }) },
        { name: "wizard_cluster_load_balancing_mode", type: "select", label: byLang({ ru: "Режим балансировки", be: "Рэжым балансавання", en: "Load balancing mode" }), options: ["performance", "memory"] },
        { name: "wizard_cluster_security_level", type: "text", label: byLang({ ru: "Уровень безопасности", be: "Узровень бяспекі", en: "Security level" }) },
        { name: "wizard_cluster_session_fault_tolerance_level", type: "text", label: byLang({ ru: "Уровень отказоустойчивости", be: "Узровень адмоваўстойлівасці", en: "Fault tolerance level" }) },
        { name: "wizard_cluster_ping_period", type: "text", label: byLang({ ru: "Период ping, мс", be: "Перыяд ping, мс", en: "Ping period, ms" }) },
        { name: "wizard_cluster_ping_timeout", type: "text", label: byLang({ ru: "Таймаут ping, мс", be: "Таймаўт ping, мс", en: "Ping timeout, ms" }) },
        { name: "wizard_cluster_restart_schedule", type: "text", label: byLang({ ru: "Расписание перезапуска", be: "Расклад перазапуску", en: "Restart schedule" }) },
        { name: "wizard_cluster_kill_problem_processes", type: "bool", label: byLang({ ru: "Завершать проблемные процессы", be: "Завяршаць праблемныя працэсы", en: "Kill problematic processes" }) },
        { name: "wizard_cluster_kill_by_memory_with_dump", type: "bool", label: byLang({ ru: "Формировать дамп при превышении памяти", be: "Ствараць дамп пры перавышэнні памяці", en: "Create dump on memory overflow" }) },
        { name: "wizard_cluster_allow_access_audit", type: "bool", label: byLang({ ru: "Разрешить аудит прав доступа", be: "Дазволіць аўдыт правоў доступу", en: "Allow access-right audit events" }) },
      ],
    },
  };

  function t(key, fallback) {
    return locale[key] || fallback;
  }

  function displayMode(mode) {
    const value = String(mode || "").trim();
    if (!value) return t("allEntities", "All");
    return (locale.entityNames && locale.entityNames[value]) || value.charAt(0).toUpperCase() + value.slice(1);
  }

  function displayRisk(value) {
    switch (String(value || "").toLowerCase()) {
      case "low":
        return t("riskLow", "low");
      case "medium":
        return t("riskMedium", "medium");
      case "high":
        return t("riskHigh", "high");
      case "critical":
        return t("riskCritical", "critical");
      default:
        return value || "";
    }
  }

  function localizedOperationTitle(item) {
    const exact = {
      be: {
        "rac.cluster.list": "Спіс кластараў",
        "rac.cluster.info": "Інфармацыя пра кластар",
        "rac.cluster.admin.list": "Адміністратары кластара",
        "rac.cluster.update": "Абнаўленне кластара",
        "rac.infobase.summary.list": "Спіс інфабаз",
        "rac.infobase.info": "Інфармацыя пра інфабазу",
        "rac.infobase.create": "Стварыць інфабазу",
        "rac.infobase.drop": "Выдаліць інфабазу",
        "rac.session.list": "Спіс сеансаў",
        "rac.session.info": "Інфармацыя пра сеанс",
        "rac.session.terminate": "Завяршыць сеанс",
        "rac.session.interrupt-call": "Перапыніць серверны выклік",
        "rac.connection.list": "Спіс злучэнняў",
        "rac.connection.info": "Інфармацыя пра злучэнне",
        "rac.connection.disconnect": "Адключыць злучэнне",
        "rac.server.list": "Спіс сервераў",
        "rac.server.info": "Інфармацыя пра сервер",
        "rac.server.update": "Абнаўленне сервера"
      },
      en: {
        "rac.cluster.list": "Cluster list",
        "rac.cluster.info": "Cluster information",
        "rac.cluster.admin.list": "Cluster administrators",
        "rac.cluster.update": "Cluster update",
        "rac.infobase.summary.list": "Infobase list",
        "rac.infobase.info": "Infobase information",
        "rac.infobase.create": "Create infobase",
        "rac.infobase.drop": "Delete infobase",
        "rac.session.list": "Session list",
        "rac.session.info": "Session information",
        "rac.session.terminate": "Terminate session",
        "rac.session.interrupt-call": "Interrupt server call",
        "rac.connection.list": "Connection list",
        "rac.connection.info": "Connection information",
        "rac.connection.disconnect": "Disconnect connection",
        "rac.server.list": "Server list",
        "rac.server.info": "Server information",
        "rac.server.update": "Server update"
      }
    };
    const localized = exact[lang] && exact[lang][item.id];
    if (localized) return localized;
    if (lang === "ru") return item.title;
    return `${displayMode(item.mode)}: ${quickActionSuffixLabel(item.id) || item.id}`;
  }

  function localizedOperationDescription(item) {
    const exact = {
      be: {
        "rac.cluster.list": "Атрымаць спіс кластараў",
        "rac.cluster.info": "Атрымаць падрабязнасці пра кластар",
        "rac.cluster.admin.list": "Праглядзець адміністратараў кластара",
        "rac.cluster.update": "Змяніць параметры кластара",
        "rac.infobase.summary.list": "Праглядзець кароткі спіс інфабаз",
        "rac.infobase.info": "Атрымаць дэталі інфабазы",
        "rac.infobase.create": "Стварыць новую інфабазу",
        "rac.infobase.drop": "Выдаліць інфабазу",
        "rac.session.list": "Праглядзець актыўныя сеансы",
        "rac.session.info": "Атрымаць дэталі сеансу",
        "rac.session.terminate": "Прымусова завяршыць сеанс",
        "rac.session.interrupt-call": "Перапыніць бягучы серверны выклік",
        "rac.connection.list": "Праглядзець злучэнні",
        "rac.connection.info": "Атрымаць дэталі злучэння",
        "rac.connection.disconnect": "Прымусова адключыць злучэнне",
        "rac.server.list": "Праглядзець серверы кластара",
        "rac.server.info": "Атрымаць дэталі сервера",
        "rac.server.update": "Змяніць параметры сервера"
      },
      en: {
        "rac.cluster.list": "Get the list of clusters",
        "rac.cluster.info": "Get cluster details",
        "rac.cluster.admin.list": "View cluster administrators",
        "rac.cluster.update": "Update cluster settings",
        "rac.infobase.summary.list": "View the infobase list",
        "rac.infobase.info": "Get infobase details",
        "rac.infobase.create": "Create a new infobase",
        "rac.infobase.drop": "Delete an infobase",
        "rac.session.list": "View active sessions",
        "rac.session.info": "Get session details",
        "rac.session.terminate": "Force terminate a session",
        "rac.session.interrupt-call": "Interrupt the current server call",
        "rac.connection.list": "View connections",
        "rac.connection.info": "Get connection details",
        "rac.connection.disconnect": "Force disconnect a connection",
        "rac.server.list": "View cluster servers",
        "rac.server.info": "Get server details",
        "rac.server.update": "Update server settings"
      }
    };
    const localized = exact[lang] && exact[lang][item.id];
    if (localized) return localized;
    if (lang === "ru") return item.description;
    return localizedOperationTitle(item);
  }

  function entityModes() {
    return [...new Set(operations.map((item) => item.mode).filter(Boolean))];
  }

  function operationById(id) {
    return operations.find((item) => item.id === id) || null;
  }

  function resolvedOperatorFlows() {
    const availableIds = new Set(operations.map((item) => item.id));
    return operatorFlowDefinitions
      .map((flow) => {
        const allowedSteps = [];
        const skippedOptional = [];
        let missingRequired = false;
        flow.steps.forEach((step) => {
          if (availableIds.has(step.operationId)) {
            allowedSteps.push(step);
          } else if (step.required) {
            missingRequired = true;
          } else {
            skippedOptional.push(step);
          }
        });
        if (missingRequired || !allowedSteps.length) return null;
        return { ...flow, steps: allowedSteps, skippedOptional };
      })
      .filter(Boolean);
  }

  function availablePlaybooksForEntity(mode) {
    return playbookDefinitions[mode] || [];
  }

  function ensurePlaybookSelection() {
    const items = availablePlaybooksForEntity(activeEntity);
    if (!items.length) {
      activePlaybook = "";
      return;
    }
    if (!items.some((item) => item.id === activePlaybook)) {
      activePlaybook = items[0].id;
    }
  }

  function playbookForOperation(mode, operationId) {
    return availablePlaybooksForEntity(mode).find((item) => item.opIds.includes(operationId)) || null;
  }

  function filteredOperationsByEntity() {
    const base = activeEntity ? operations.filter((item) => item.mode === activeEntity) : operations;
    if (!activeEntity || !activePlaybook) return base;
    const playbook = availablePlaybooksForEntity(activeEntity).find((item) => item.id === activePlaybook);
    if (!playbook) return base;
    return base.filter((item) => playbook.opIds.includes(item.id));
  }

  function currentOperation() {
    return operations.find((item) => item.id === (select ? select.value : ""));
  }

  function importantFieldNames(mode) {
    const common = ["host", "admin_port"];
    const byMode = {
      cluster: ["cluster", "cluster-host", "cluster-port", "cluster-user"],
      infobase: ["infobase", "name", "infobase-user", "db-server", "db-name"],
      session: ["session", "infobase", "name"],
      connection: ["connection", "infobase", "name"],
      server: ["server", "agent-host", "agent-port"],
      manager: ["manager"],
      process: ["process"],
      rule: ["server", "rule"],
      counter: ["counter"],
      limit: ["limit"],
      profile: ["name"],
      "binary-data-storage": ["binary-data-storage"],
    };
    return [...common, ...(byMode[mode] || [])];
  }

  function splitParams(operation) {
    const preferred = importantFieldNames(operation.mode);
    const smartDefinition = smartWizardDefinitionFor(operation);
    const hiddenParams = new Set((smartDefinition && smartDefinition.hideParams) || []);
    const primary = [];
    const secondary = [];
    operation.params.forEach((param) => {
      if (hiddenParams.has(param.name)) return;
      const isPrimary = param.required || preferred.includes(param.name);
      (isPrimary ? primary : secondary).push(param);
    });
    return { primary, secondary };
  }

  function textInputAutocomplete(name, type) {
    if (type === "password") return 'autocomplete="new-password"';
    if (isSensitiveField(name)) return 'autocomplete="off" autocapitalize="off" spellcheck="false"';
    return "";
  }

  function fieldHtml(param) {
    const safeName = param.name;
    const label = param.label || param.name;
    if (param.type === "bool") {
      return `<label><span>${label}</span><input type="checkbox" name="${safeName}" ${textInputAutocomplete(safeName, "checkbox")}></label>`;
    }
    if (param.type === "select") {
      const options = (param.options || []).map((item) => `<option value="${item}">${item}</option>`).join("");
      return `<label>${label}<select name="${safeName}" ${param.required ? "required" : ""}><option value=""></option>${options}</select></label>`;
    }
    const type = param.type === "password" ? "password" : "text";
    const value = Object.prototype.hasOwnProperty.call(defaults, safeName) ? `value="${defaults[safeName]}"` : "";
    return `<label>${label}<input name="${safeName}" type="${type}" ${value} ${param.required ? "required" : ""} ${textInputAutocomplete(safeName, type)}></label>`;
  }

  function isSensitiveField(name) {
    const lower = String(name || "").toLowerCase();
    return lower.includes("pwd") || lower.includes("password") || ["cluster-user", "infobase-user", "db-user", "agent-user", "os-user"].includes(lower);
  }

  function syncOperationOptions() {
    if (!select) return;
    const currentValue = select.value;
    const filtered = filteredOperationsByEntity();
    select.innerHTML = '<option value="">--</option>';
    filtered.forEach((item) => {
      const option = document.createElement("option");
      option.value = item.id;
      option.textContent = `${item.id} - ${localizedOperationTitle(item)}`;
      select.appendChild(option);
    });
    if (filtered.some((item) => item.id === currentValue)) {
      select.value = currentValue;
    }
  }

  function quickActionSuffixLabel(id) {
    if (id.endsWith(".list")) return t("quickList", "List");
    if (id.endsWith(".info")) return t("quickInfo", "Info");
    if (id.endsWith(".create") || id.endsWith(".insert")) return t("quickCreate", "Create");
    if (id.endsWith(".update")) return t("quickUpdate", "Update");
    if (id.endsWith(".remove") || id.endsWith(".drop") || id.endsWith(".terminate") || id.endsWith(".disconnect")) return t("quickRemove", "Remove");
    return "";
  }

  function renderQuickActions() {
    if (!quickActions) return;
    quickActions.innerHTML = "";
    filteredOperationsByEntity()
      .filter((item) => /\.(list|info|create|insert|update|remove|drop|terminate|disconnect)$/.test(item.id))
      .slice(0, 6)
      .forEach((item) => {
        const button = document.createElement("button");
        button.type = "button";
        button.className = "pill-button";
        button.textContent = quickActionSuffixLabel(item.id) || localizedOperationTitle(item);
        button.addEventListener("click", () => {
          select.value = item.id;
          render();
        });
        quickActions.appendChild(button);
      });
  }

  function currentResolvedFlow() {
    return resolvedOperatorFlows().find((item) => item.id === activeFlowId) || null;
  }

  function suggestedFieldsForOperation(operationId) {
    const mapping = {
      "rac.cluster.info": ["cluster"],
      "rac.infobase.update": ["cluster", "infobase", "name"],
      "rac.infobase.drop": ["cluster", "infobase", "name"],
      "rac.session.list": ["cluster", "infobase", "name"],
      "rac.session.info": ["cluster", "session"],
      "rac.session.terminate": ["cluster", "session"],
      "rac.connection.list": ["cluster", "infobase", "name"],
      "rac.connection.info": ["cluster", "connection"],
      "rac.connection.disconnect": ["cluster", "process", "connection"],
      "rac.cluster.update": ["cluster"],
    };
    return mapping[operationId] || [];
  }

  function tryPopulateFromLatestResult(operationId) {
    const records = [...structuredResultRecords(), ...parseResultRecords(latestResultText())];
    if (!records.length) return;
    const keys = suggestedFieldsForOperation(operationId);
    keys.forEach((key) => {
      const uniqueValues = [...new Set(records.map((record) => record[key]).filter(Boolean))];
      if (uniqueValues.length === 1) {
        applyResultValue(key, uniqueValues[0]);
      }
    });
  }

  function activateFlowStep(flowId, stepIndex) {
    const flow = resolvedOperatorFlows().find((item) => item.id === flowId);
    if (!flow || !flow.steps[stepIndex] || !select) return;
    activeFlowId = flowId;
    activeFlowStepIndex = stepIndex;
    const op = operationById(flow.steps[stepIndex].operationId);
    if (op && op.mode) {
      activeEntity = op.mode;
      const matchingPlaybook = playbookForOperation(op.mode, op.id);
      activePlaybook = matchingPlaybook ? matchingPlaybook.id : activePlaybook;
      syncOperationOptions();
      select.value = op.id;
      render();
      tryPopulateFromLatestResult(op.id);
      render();
    }
  }

  function renderOperatorFlows() {
    if (!operatorFlowHost) return;
    const flows = resolvedOperatorFlows();
    if (!flows.length) {
      operatorFlowHost.innerHTML = "";
      activeFlowId = "";
      activeFlowStepIndex = 0;
      return;
    }
    const selectedFlow = flows.find((flow) => flow.id === activeFlowId) || flows[0];
    if (!flows.some((flow) => flow.id === activeFlowId)) {
      activeFlowId = "";
      activeFlowStepIndex = 0;
    }
    const activeFlow = currentResolvedFlow() || selectedFlow;
    if (activeFlow && activeFlowStepIndex >= activeFlow.steps.length) {
      activeFlowStepIndex = activeFlow.steps.length - 1;
    }
    operatorFlowHost.innerHTML = `
      <section class="smart-wizard-card operator-flow-card">
        <div class="entity-panel-head">
          <span class="section-label">${byLang({ ru: "Операторские цепочки", be: "Аператарскія ланцужкі", en: "Operator flows" })}</span>
          <span class="subtle-note">${byLang({ ru: "Готовые последовательности команд с учетом ваших прав.", be: "Гатовыя паслядоўнасці каманд з улікам вашых правоў.", en: "Ready command chains that respect the current role." })}</span>
        </div>
        <div class="wizard-grid smart-preset-grid">
          ${flows.map((flow) => `
            <button type="button" class="${flow.id === (activeFlow && activeFlow.id) ? "wizard-tile is-active" : "wizard-tile"}" data-flow-id="${flow.id}">
              <strong>${flow.title}</strong>
              <span>${flow.hint}</span>
              ${flow.skippedOptional.length ? `<small>${byLang({ ru: "Некоторые шаги скрыты по роли", be: "Некаторыя крокі схаваны па ролі", en: "Some steps are hidden by role" })}</small>` : ""}
            </button>
          `).join("")}
        </div>
        ${activeFlow ? `
          <div class="flow-progress-card">
            <div class="entity-panel-head">
              <span class="section-label">${activeFlow.title}</span>
              <span class="subtle-note">${activeFlow.hint}</span>
            </div>
            <ol class="flow-step-list">
              ${activeFlow.steps.map((step, index) => `
                <li class="${index === activeFlowStepIndex ? "flow-step is-active" : "flow-step"}">
                  <button type="button" data-flow-step="${index}">
                    <strong>${index + 1}. ${step.label}</strong>
                    <span>${localizedOperationTitle(operationById(step.operationId) || { id: step.operationId, title: step.operationId, mode: activeFlow.entity, risk_level: "" })}</span>
                  </button>
                </li>
              `).join("")}
            </ol>
            <div class="action-row flow-actions">
              <button type="button" data-flow-start>${byLang({ ru: "Начать цепочку", be: "Пачаць ланцужок", en: "Start flow" })}</button>
              <button type="button" data-flow-next ${activeFlowStepIndex >= activeFlow.steps.length - 1 ? "disabled" : ""}>${byLang({ ru: "Следующий шаг", be: "Наступны крок", en: "Next step" })}</button>
              <button type="button" data-flow-reset>${byLang({ ru: "Сбросить", be: "Скінуць", en: "Reset" })}</button>
            </div>
            ${activeFlow.skippedOptional.length ? `<p class="subtle-note">${byLang({ ru: "Шаги, скрытые из-за роли", be: "Крокі, схаваныя праз ролю", en: "Steps omitted because of the current role" })}: ${activeFlow.skippedOptional.map((step) => step.label).join(", ")}</p>` : ""}
          </div>
        ` : ""}
      </section>
    `;
    operatorFlowHost.querySelectorAll("[data-flow-id]").forEach((node) => {
      node.addEventListener("click", () => {
        activeFlowId = node.getAttribute("data-flow-id") || "";
        activeFlowStepIndex = 0;
        render();
      });
    });
    const startButton = operatorFlowHost.querySelector("[data-flow-start]");
    if (startButton && activeFlow) {
      startButton.addEventListener("click", () => activateFlowStep(activeFlow.id, 0));
    }
    const nextButton = operatorFlowHost.querySelector("[data-flow-next]");
    if (nextButton && activeFlow) {
      nextButton.addEventListener("click", () => activateFlowStep(activeFlow.id, Math.min(activeFlowStepIndex + 1, activeFlow.steps.length - 1)));
    }
    const resetButton = operatorFlowHost.querySelector("[data-flow-reset]");
    if (resetButton) {
      resetButton.addEventListener("click", () => {
        activeFlowStepIndex = 0;
        render();
      });
    }
    operatorFlowHost.querySelectorAll("[data-flow-step]").forEach((node) => {
      node.addEventListener("click", () => {
        const stepIndex = Number(node.getAttribute("data-flow-step"));
        if (activeFlow) activateFlowStep(activeFlow.id, stepIndex);
      });
    });
  }

  function renderGuidedScenarios() {
    if (!guidedScenarios) return;
    guidedScenarios.innerHTML = "";
    const mode = activeEntity || "cluster";
    const items = (scenarioDefinitions[mode] || scenarioDefinitions.cluster || [])
      .map((id) => operations.find((item) => item.id === id))
      .filter(Boolean)
      .slice(0, 4);
    items.forEach((item) => {
      const tile = document.createElement("button");
      tile.type = "button";
      tile.className = "wizard-tile";
      tile.innerHTML = `<strong>${localizedOperationTitle(item)}</strong><span>${localizedOperationDescription(item)}</span>`;
      tile.addEventListener("click", () => {
        activeEntity = item.mode || activeEntity;
        const matchingPlaybook = playbookForOperation(activeEntity, item.id);
        activePlaybook = matchingPlaybook ? matchingPlaybook.id : activePlaybook;
        renderEntityPills(entityFilters, setActiveEntity);
        renderEntityPills(catalogEntityFilters, setCatalogEntity);
        syncOperationOptions();
        renderQuickActions();
        renderGuidedScenarios();
        if (select) {
          select.value = item.id;
        }
        render();
        const firstEmpty = primaryFields ? [...primaryFields.querySelectorAll("input, select, textarea")].find((field) => {
          if (field.type === "hidden" || field.type === "checkbox" || field.disabled) return false;
          return !field.value;
        }) : null;
        if (firstEmpty) firstEmpty.focus();
      });
      guidedScenarios.appendChild(tile);
    });
  }

  function favoriteCountForOperation(operationId) {
    if (!operationId) return 0;
    return favorites.filter((item) => (item.OperationID || item.operation_id) === operationId).length;
  }

  function renderJourneySummary(operation) {
    if (!journeySummary) return;
    const selectedProfile = profiles.find((item) => String(item.ID || item.id) === (profileSelect ? profileSelect.value : ""));
    const profileText = selectedProfile ? `${selectedProfile.Name || selectedProfile.name} (${selectedProfile.Host || selectedProfile.host}:${selectedProfile.Port || selectedProfile.port})` : t("summaryPending", "not selected");
    const favoriteText = operation ? String(favoriteCountForOperation(operation.id)) : t("summaryNotSaved", "not saved");
    const selectedTrack = availablePlaybooksForEntity(activeEntity).find((item) => item.id === activePlaybook);
    const trackText = selectedTrack ? t(selectedTrack.labelKey, selectedTrack.id) : t("summaryPending", "not selected");
    const steps = [
      t("stepEntity", "Entity"),
      t("stepAction", "Action"),
      t("stepParams", "Parameters"),
      t("stepRun", "Run"),
    ];
    const activeStep = operation ? 2 : activeEntity ? 1 : 0;

    journeySummary.innerHTML = `
      <div class="detail-card">
        <span class="section-label">${t("summaryTitle", "Execution flow")}</span>
        <div class="journey-steps">
          ${steps.map((label, index) => `<span class="${index <= activeStep ? "journey-step is-active" : "journey-step"}">${index + 1}. ${label}</span>`).join("")}
        </div>
        <div class="journey-grid">
          <span><strong>${t("summaryEntity", "Entity")}:</strong> ${activeEntity ? displayMode(activeEntity) : t("summaryPending", "not selected")}</span>
          <span><strong>${t("summaryAction", "Action")}:</strong> ${operation ? localizedOperationTitle(operation) : t("summaryPending", "not selected")}</span>
          <span><strong>${t("summaryTrack", "Work track")}:</strong> ${trackText}</span>
          <span><strong>${t("summaryProfile", "Profile")}:</strong> ${profileText}</span>
          <span><strong>${t("summaryFavorite", "Favorite")}:</strong> ${favoriteText}</span>
        </div>
      </div>
    `;
  }

  function renderPlaybooks() {
    if (!playbookList) return;
    if (playbookLabel) playbookLabel.textContent = t("playbookTitle", "Work track");
    if (playbookHint) playbookHint.textContent = t("playbookHint", "Pick the task type first, then choose a concrete action");
    ensurePlaybookSelection();
    const items = availablePlaybooksForEntity(activeEntity);
    playbookList.innerHTML = "";
    items.forEach((item) => {
      const button = document.createElement("button");
      button.type = "button";
      button.className = item.id === activePlaybook ? "playbook-card is-active" : "playbook-card";
      button.innerHTML = `<strong>${t(item.labelKey, item.id)}</strong><span>${t(item.hintKey, item.id)}</span>`;
      button.addEventListener("click", () => {
        activePlaybook = item.id;
        syncOperationOptions();
        const filtered = filteredOperationsByEntity();
        if (select && filtered.length && !filtered.some((entry) => entry.id === select.value)) {
          select.value = filtered[0].id;
        }
        render();
      });
      playbookList.appendChild(button);
    });
  }

  function renderHumanActionList() {
    if (!humanActionList) return;
    if (friendlyActionLabel) friendlyActionLabel.textContent = t("selectActionTitle", "Action");
    if (friendlyActionHint) friendlyActionHint.textContent = t("selectActionHint", "Choose a human-friendly action and the technical command will be selected automatically");
    if (technicalCommandLabel) technicalCommandLabel.textContent = t("technicalCommand", "Technical command");
    if (technicalCommandHint) technicalCommandHint.textContent = t("technicalCommandHint", "Kept for precise control and full rac / ras compatibility");

    const operation = currentOperation();
    const filtered = filteredOperationsByEntity().slice(0, 12);
    humanActionList.innerHTML = "";
    filtered.forEach((item) => {
      const button = document.createElement("button");
      button.type = "button";
      button.className = item.id === (operation && operation.id) ? "action-card is-active" : "action-card";
      button.innerHTML = `
        <strong>${localizedOperationTitle(item)}</strong>
        <span>${localizedOperationDescription(item)}</span>
        <small>${displayMode(item.mode)} · ${displayRisk(item.risk_level)}</small>
      `;
      button.addEventListener("click", () => {
        if (select) {
          select.value = item.id;
        }
        render();
        const firstEmpty = primaryFields ? [...primaryFields.querySelectorAll("input, select, textarea")].find((field) => {
          if (field.type === "hidden" || field.type === "checkbox" || field.disabled) return false;
          return !field.value;
        }) : null;
        if (firstEmpty) firstEmpty.focus();
      });
      humanActionList.appendChild(button);
    });
  }

  function renderFieldGuide(operation) {
    if (!fieldGuide) return;
    if (fieldGuideLabel) fieldGuideLabel.textContent = t("fieldGuideTitle", "What to prepare");
    if (fieldGuideHint) fieldGuideHint.textContent = t("fieldGuideHint", "The wizard highlights the key fields for the current action");
    if (!operation) {
      fieldGuide.innerHTML = `<p class="subtle-note">${t("fieldGuideEmpty", "Choose an action and the wizard will show the key fields.")}</p>`;
      return;
    }
    const grouped = splitParams(operation);
    const primaryItems = grouped.primary.map((param) => `<li>${param.label || param.name}</li>`).join("");
    const secondaryItems = grouped.secondary.slice(0, 4).map((param) => `<li>${param.label || param.name}</li>`).join("");
    fieldGuide.innerHTML = `
      <div class="field-guide-section">
        <strong>${t("fieldGuidePrimary", "Primary fields")}</strong>
        <ul>${primaryItems || `<li>${t("summaryPending", "not selected")}</li>`}</ul>
      </div>
      <div class="field-guide-section">
        <strong>${t("fieldGuideAdvanced", "Advanced parameters")}</strong>
        <ul>${secondaryItems || `<li>${t("actionReady", "Ready to execute")}</li>`}</ul>
      </div>
    `;
  }

  function renderEntityPills(containerNode, onClick) {
    if (!containerNode) return;
    containerNode.innerHTML = "";
    ["", ...entityModes()].forEach((mode) => {
      const button = document.createElement("button");
      button.type = "button";
      button.className = mode === activeEntity ? "pill-button is-active" : "pill-button";
      button.textContent = displayMode(mode);
      button.addEventListener("click", () => onClick(mode));
      containerNode.appendChild(button);
    });
  }

  function renderScenarioAssistant(operation) {
    if (!wizardAssistant) return;
    wizardAssistant.innerHTML = "";
    if (!operation) return;
    const grouped = splitParams(operation);
    const keyFields = grouped.primary.map((param) => param.label || param.name).slice(0, 4);
    wizardAssistant.innerHTML = `
      <div class="section-label">${t("scenarioSteps", "Guided scenario")}</div>
      <ol>
        <li>${displayMode(operation.mode)}: ${localizedOperationTitle(operation)}</li>
        <li>${localizedOperationDescription(operation)}</li>
        <li>${keyFields.length ? keyFields.join(", ") : t("primaryFields", "Primary fields")}</li>
        <li>${t("advancedFields", "Additional parameters")}</li>
      </ol>
    `;
  }

  function applyValues(values) {
    if (!values) return;
    Object.entries(values).forEach(([key, value]) => {
      if (key === "operation_id") return;
      if (isSensitiveField(key)) return;
      const field = document.querySelector(`[name="${key}"]`);
      if (!field) return;
      if (field.type === "checkbox") {
        field.checked = value === "on" || value === "true" || value === "1";
      } else {
        field.value = value;
      }
    });
  }

  function applyProfileDefaults() {
    if (!profileSelect) return;
    const selected = profiles.find((item) => String(item.ID || item.id) === profileSelect.value);
    const hostInput = document.querySelector('input[name="host"]');
    const portInput = document.querySelector('input[name="admin_port"]');
    if (selected && hostInput) hostInput.value = selected.Host || selected.host || hostInput.value;
    if (selected && portInput) portInput.value = String(selected.Port || selected.port || portInput.value);
    if (connectionProfileIdInput) connectionProfileIdInput.value = profileSelect.value || "";
    if (selected && favoriteConnection) favoriteConnection.value = String(selected.ID || selected.id || "");
  }

  function selectedFavoriteObject() {
    if (!favoriteSelect) return null;
    return favorites.find((item) => String(item.ID || item.id) === favoriteSelect.value) || null;
  }

  function refreshFavoriteOptions() {
    if (!favoriteSelect) return;
    const previous = favoriteSelect.value;
    const operationId = select ? select.value : "";
    const filtered = operationId ? favorites.filter((item) => (item.OperationID || item.operation_id) === operationId) : [];
    favoriteSelect.innerHTML = '<option value="">--</option>';
    filtered.forEach((item) => {
      const option = document.createElement("option");
      option.value = String(item.ID || item.id);
      option.textContent = item.Name || item.name || option.value;
      favoriteSelect.appendChild(option);
    });
    if (filtered.some((item) => String(item.ID || item.id) === previous)) {
      favoriteSelect.value = previous;
    }
    if (favoriteNote) {
      favoriteNote.textContent = operationId ? (filtered.length ? t("selectFavorite", "Apply favorite") : t("noFavorites", "No favorites saved yet")) : t("selectOperationForFavorites", "Select an operation to see favorites");
    }
    if (favoriteIdInput) {
      favoriteIdInput.value = favoriteSelect.value || "";
    }
    if (deleteFavoriteButton) {
      deleteFavoriteButton.disabled = !favoriteSelect.value;
    }
  }

  function renderOperationDetails(operation) {
    if (!operationDetails) return;
    operationDetails.innerHTML = "";
    if (!operation) return;
    operationDetails.innerHTML = `
      <div class="detail-card">
        <span class="section-label">${t("operationDetails", "Operation details")}</span>
        <div class="detail-grid">
          <span><strong>${t("labelId", "ID")}:</strong> ${operation.id}</span>
          <span><strong>${t("labelMode", "Mode")}:</strong> ${displayMode(operation.mode)}</span>
          <span><strong>${t("labelUtility", "Utility")}:</strong> ${operation.utility}</span>
          <span><strong>${t("labelRisk", "Risk")}:</strong> ${displayRisk(operation.risk_level)}</span>
        </div>
        <p class="subtle-note">${t("technicalCommandHint", "Kept for precise control and full rac / ras compatibility")}</p>
      </div>
    `;
  }

  function latestResultText() {
    const outputs = document.querySelectorAll(".execute-result-card .terminal-output");
    if (!outputs.length) return "";
    return outputs[0].textContent || "";
  }

  function structuredResultRecords() {
    const rows = document.querySelectorAll(".parsed-result tbody tr");
    if (!rows.length) return [];
    const record = {};
    rows.forEach((row) => {
      const cells = row.querySelectorAll("td");
      if (cells.length < 2) return;
      const key = (cells[0].textContent || "").trim();
      const value = (cells[1].textContent || "").trim();
      if (!key) return;
      record[key] = value;
    });
    return Object.keys(record).length ? [record] : [];
  }

  function parseResultRecords(raw) {
    const lines = String(raw || "").split(/\r?\n/);
    const records = [];
    let current = {};
    let hasData = false;
    for (const rawLine of lines) {
      const line = rawLine.trim();
      if (!line) {
        if (hasData) {
          records.push(current);
          current = {};
          hasData = false;
        }
        continue;
      }
      const idx = line.indexOf(":");
      if (idx === -1) continue;
      const key = line.slice(0, idx).trim();
      const value = line.slice(idx + 1).trim();
      if (!key) continue;
      if (Object.prototype.hasOwnProperty.call(current, key)) {
        records.push(current);
        current = {};
      }
      current[key] = value;
      hasData = true;
    }
    if (hasData) records.push(current);
    return records;
  }

  function relevantResultKeys(operation) {
    const mode = operation ? operation.mode : "";
    const generic = ["cluster", "infobase", "name", "process", "server"];
    const byMode = {
      session: ["session", "infobase", "name", "cluster"],
      connection: ["connection", "process", "infobase", "name", "cluster"],
      infobase: ["infobase", "name", "cluster"],
      cluster: ["cluster", "name"],
      server: ["server", "cluster", "name"],
      process: ["process", "server", "cluster"],
    };
    return byMode[mode] || generic;
  }

  function applyResultValue(fieldName, value) {
    const target = document.querySelector(`[name="${fieldName}"]`);
    if (!target || !value) return;
    if (target.type === "checkbox") {
      target.checked = value === "true" || value === "on" || value === "1";
    } else {
      target.value = value;
    }
    target.dispatchEvent(new Event("input", { bubbles: true }));
    target.dispatchEvent(new Event("change", { bubbles: true }));
  }

  function renderResultContext(operation) {
    if (!resultContextHost) return;
    const records = [...structuredResultRecords(), ...parseResultRecords(latestResultText())].slice(0, 8);
    if (!operation || !records.length) {
      resultContextHost.innerHTML = "";
      return;
    }
    const keys = relevantResultKeys(operation);
    const entries = [];
    records.forEach((record, index) => {
      keys.forEach((key) => {
        if (record[key]) {
          entries.push({ key, value: record[key], record, index });
        }
      });
    });
    if (!entries.length) {
      resultContextHost.innerHTML = "";
      return;
    }
    resultContextHost.innerHTML = `
      <section class="smart-wizard-card result-context-card">
        <div class="entity-panel-head">
          <span class="section-label">${byLang({ ru: "Подставить из последнего результата", be: "Падставіць з апошняга выніку", en: "Reuse values from the latest result" })}</span>
          <span class="subtle-note">${byLang({ ru: "Быстро переносите найденные UUID и имена в следующую команду.", be: "Хутка пераносіце знойдзеныя UUID і імёны ў наступную каманду.", en: "Quickly carry UUIDs and names into the next command." })}</span>
        </div>
        <div class="result-context-grid">
          ${entries.map((entry) => `
            <button type="button" class="pill-button result-context-chip" data-result-index="${entry.index}" data-result-field="${entry.key}">
              <strong>${entry.key}</strong>
              <span>${entry.value}</span>
            </button>
          `).join("")}
        </div>
      </section>
    `;
    resultContextHost.querySelectorAll("[data-result-field]").forEach((node) => {
      node.addEventListener("click", () => {
        const key = node.getAttribute("data-result-field");
        const index = Number(node.getAttribute("data-result-index"));
        const source = entries.find((item) => item.index === index && item.key === key);
        const value = source ? source.value : "";
        applyResultValue(key, value);
      });
    });
  }

  function smartWizardDefinitionFor(operation) {
    if (!operation) return null;
    return smartWizardDefinitions[operation.id] || null;
  }

  function wizardFieldValue(name, type) {
    const field = document.querySelector(`[name="${name}"]`);
    if (!field) return "";
    if (type === "bool") {
      return field.checked ? "true" : "false";
    }
    return field.value || "";
  }

  function normalizeWizardDate(value) {
    const trimmed = String(value || "").trim();
    if (!trimmed) return "";
    if (/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}$/.test(trimmed)) {
      return `${trimmed}:00`;
    }
    return trimmed;
  }

  function buildSmartWizardArgs(definition) {
    if (!definition) return [];
    if (definition.builder === "cluster_update") {
      const mappings = [
        ["wizard_cluster_name", "--name"],
        ["wizard_cluster_load_balancing_mode", "--load-balancing-mode"],
        ["wizard_cluster_security_level", "--security-level"],
        ["wizard_cluster_session_fault_tolerance_level", "--session-fault-tolerance-level"],
        ["wizard_cluster_ping_period", "--ping-period"],
        ["wizard_cluster_ping_timeout", "--ping-timeout"],
        ["wizard_cluster_restart_schedule", "--restart-schedule"],
      ];
      const args = [];
      mappings.forEach(([fieldName, flag]) => {
        const value = wizardFieldValue(fieldName, "text");
        if (value) args.push(`${flag}=${value}`);
      });
      if (wizardFieldValue("wizard_cluster_kill_problem_processes", "bool") === "true") args.push("--kill-problem-processes=yes");
      if (wizardFieldValue("wizard_cluster_kill_problem_processes", "bool") === "false") args.push("--kill-problem-processes=no");
      if (wizardFieldValue("wizard_cluster_kill_by_memory_with_dump", "bool") === "true") args.push("--kill-by-memory-with-dump=yes");
      if (wizardFieldValue("wizard_cluster_kill_by_memory_with_dump", "bool") === "false") args.push("--kill-by-memory-with-dump=no");
      if (wizardFieldValue("wizard_cluster_allow_access_audit", "bool") === "true") args.push("--allow-access-right-audit-events-recording=yes");
      if (wizardFieldValue("wizard_cluster_allow_access_audit", "bool") === "false") args.push("--allow-access-right-audit-events-recording=no");
      return args;
    }
    if (definition.builder !== "infobase_lock") return [];
    const values = {};
    definition.fields.forEach((field) => {
      values[field.name] = wizardFieldValue(field.name, field.type);
    });
    const args = [];
    args.push(`--sessions-deny=${values.wizard_sessions_deny === "true" ? "on" : "off"}`);
    args.push(`--scheduled-jobs-deny=${values.wizard_scheduled_jobs_deny === "true" ? "on" : "off"}`);
    if (values.wizard_denied_message) args.push(`--denied-message=${values.wizard_denied_message}`);
    if (values.wizard_permission_code) args.push(`--permission-code=${values.wizard_permission_code}`);
    if (values.wizard_denied_parameter) args.push(`--denied-parameter=${values.wizard_denied_parameter}`);
    if (values.wizard_denied_from) args.push(`--denied-from=${normalizeWizardDate(values.wizard_denied_from)}`);
    if (values.wizard_denied_to) args.push(`--denied-to=${normalizeWizardDate(values.wizard_denied_to)}`);
    return args;
  }

  function updateSmartWizardState(operation) {
    const definition = smartWizardDefinitionFor(operation);
    if (!wizardArgsJsonInput) return;
    if (!definition) {
      wizardArgsJsonInput.value = "";
      return;
    }
    const args = buildSmartWizardArgs(definition);
    wizardArgsJsonInput.value = JSON.stringify(args);
    const preview = document.getElementById("smart_wizard_preview");
    if (preview) {
      preview.textContent = args.length ? args.join(" ") : "—";
    }
  }

  function applySmartWizardPreset(definition, preset) {
    if (!definition || !preset) return;
    activeSmartWizardPreset = preset.id;
    Object.entries(preset.values || {}).forEach(([name, value]) => {
      const field = document.querySelector(`[name="${name}"]`);
      if (!field) return;
      if (field.type === "checkbox") {
        field.checked = value === "true";
      } else {
        field.value = value;
      }
    });
    smartWizardHost.querySelectorAll("[data-smart-preset]").forEach((node) => {
      node.classList.toggle("is-active", node.getAttribute("data-smart-preset") === preset.id);
    });
    updateSmartWizardState(currentOperation());
  }

  function smartWizardFieldHtml(field) {
    if (field.type === "bool") {
      return `<label class="smart-switch"><input type="checkbox" name="${field.name}"><span>${field.label}</span></label>`;
    }
    if (field.type === "select") {
      const options = (field.options || []).map((item) => `<option value="${item}">${item}</option>`).join("");
      return `<label>${field.label}<select name="${field.name}"><option value=""></option>${options}</select></label>`;
    }
    const inputType = field.type === "password" ? "password" : field.type;
    const autoComplete = inputType === "password" ? 'autocomplete="new-password"' : 'autocomplete="off"';
    return `<label>${field.label}<input name="${field.name}" type="${inputType}" placeholder="${field.placeholder || ""}" ${autoComplete}></label>`;
  }

  function renderSmartWizard(operation) {
    if (!smartWizardHost) return;
    const definition = smartWizardDefinitionFor(operation);
    if (!definition) {
      smartWizardHost.innerHTML = "";
      activeSmartWizardPreset = "";
      updateSmartWizardState(null);
      return;
    }
    if (!definition.presets.some((preset) => preset.id === activeSmartWizardPreset)) {
      activeSmartWizardPreset = "";
    }
    smartWizardHost.innerHTML = `
      <section class="smart-wizard-card">
        <div class="entity-panel-head">
          <span class="section-label">${definition.title}</span>
          <span class="subtle-note">${definition.hint}</span>
        </div>
        <div class="smart-wizard-section">
          <div class="section-label">${definition.presetsTitle}</div>
          <div class="wizard-grid smart-preset-grid">
            ${definition.presets.map((preset) => `
              <button type="button" class="${preset.id === activeSmartWizardPreset ? "wizard-tile is-active" : "wizard-tile"}" data-smart-preset="${preset.id}">
                <strong>${preset.title}</strong>
                <span>${preset.hint}</span>
              </button>
            `).join("")}
          </div>
        </div>
        <div class="smart-wizard-section">
          <div class="section-label">${definition.fieldsTitle}</div>
          <div class="dynamic-fields smart-wizard-fields">
            ${definition.fields.map((field) => smartWizardFieldHtml(field)).join("")}
          </div>
        </div>
        ${definition.builder ? `<div class="smart-preview">
          <strong>${definition.previewTitle}</strong>
          <code id="smart_wizard_preview">—</code>
        </div>` : ""}
      </section>
    `;
    definition.presets.forEach((preset) => {
      const button = smartWizardHost.querySelector(`[data-smart-preset="${preset.id}"]`);
      if (button) {
        button.addEventListener("click", () => applySmartWizardPreset(definition, preset));
      }
    });
    definition.fields.forEach((field) => {
      const input = smartWizardHost.querySelector(`[name="${field.name}"]`);
      if (!input) return;
      const stored = formValues[field.name];
      if (input.type === "checkbox") {
        input.checked = stored === "true" || stored === "on";
      } else if (Object.prototype.hasOwnProperty.call(formValues, field.name)) {
        input.value = stored;
      }
      input.addEventListener("input", () => updateSmartWizardState(operation));
      input.addEventListener("change", () => updateSmartWizardState(operation));
    });
    updateSmartWizardState(operation);
  }

  function renderFields(operation) {
    if (!primaryFields || !advancedFields) return;
    primaryFields.innerHTML = "";
    advancedFields.innerHTML = "";
    if (advancedFieldsBox) {
      advancedFieldsBox.hidden = true;
      advancedFieldsBox.open = false;
    }
    if (!operation) return;

    const intro = document.createElement("div");
    intro.className = "alert";
    intro.textContent = `${localizedOperationTitle(operation)}. ${t("risk", "Risk")}: ${displayRisk(operation.risk_level)}. ${localizedOperationDescription(operation)}`;
    primaryFields.appendChild(intro);

    const grouped = splitParams(operation);
    grouped.primary.forEach((param) => primaryFields.insertAdjacentHTML("beforeend", fieldHtml(param)));
    grouped.secondary.forEach((param) => advancedFields.insertAdjacentHTML("beforeend", fieldHtml(param)));

    if (advancedFieldsBox) {
      advancedFieldsBox.hidden = grouped.secondary.length === 0;
    }
  }

  function render() {
    if (!select || !primaryFields || !advancedFields) return;
    const operation = currentOperation();
    if (operation && operation.mode && operation.mode !== activeEntity) {
      activeEntity = operation.mode;
    }
    if (operation) {
      const matchingPlaybook = playbookForOperation(operation.mode, operation.id);
      if (matchingPlaybook && matchingPlaybook.id !== activePlaybook) {
        activePlaybook = matchingPlaybook.id;
      }
    } else {
      ensurePlaybookSelection();
    }
    refreshFavoriteOptions();
    renderPlaybooks();
    renderHumanActionList();
    renderOperatorFlows();
    renderJourneySummary(operation);
    renderScenarioAssistant(operation);
    renderOperationDetails(operation);
    renderResultContext(operation);
    renderFieldGuide(operation);
    renderSmartWizard(operation);
    renderFields(operation);
    applyProfileDefaults();
    applyValues(formValues);
    updateSmartWizardState(operation);
  }

  function applyFavorite() {
    if (!favoriteSelect || !select) return;
    const selected = selectedFavoriteObject();
    if (favoriteIdInput) favoriteIdInput.value = favoriteSelect.value || "";
    if (deleteFavoriteButton) deleteFavoriteButton.disabled = !selected;
    if (!selected) return;
    const operationId = selected.OperationID || selected.operation_id;
    if (operationId) {
      select.value = operationId;
      render();
    }
    if (profileSelect && (selected.ConnectionID || selected.connection_id)) {
      profileSelect.value = String(selected.ConnectionID || selected.connection_id);
      applyProfileDefaults();
    }
    if (favoriteNameInput) {
      favoriteNameInput.value = selected.Name || selected.name || "";
    }
    applyValues(selected.Values || selected.values || {});
    updateSmartWizardState(currentOperation());
  }

  function applyCatalogSearch() {
    if (!catalogSearch || !catalogRows.length) return;
    const query = catalogSearch.value.trim().toLowerCase();
    catalogRows.forEach((row) => {
      const haystack = (row.getAttribute("data-search") || "").toLowerCase();
      const rowMode = row.getAttribute("data-mode") || "";
      const matchesEntity = !activeEntity || rowMode === activeEntity;
      const matchesQuery = !query || haystack.includes(query);
      row.style.display = matchesEntity && matchesQuery ? "" : "none";
    });
  }

  function setActiveEntity(mode) {
    activeEntity = mode;
    ensurePlaybookSelection();
    renderEntityPills(entityFilters, setActiveEntity);
    renderEntityPills(catalogEntityFilters, setCatalogEntity);
    syncOperationOptions();
    renderQuickActions();
    renderGuidedScenarios();
    render();
    applyCatalogSearch();
  }

  function setCatalogEntity(mode) {
    activeEntity = mode;
    ensurePlaybookSelection();
    renderEntityPills(entityFilters, setActiveEntity);
    renderEntityPills(catalogEntityFilters, setCatalogEntity);
    syncOperationOptions();
    renderQuickActions();
    renderGuidedScenarios();
    applyCatalogSearch();
  }

  if (select && primaryFields && advancedFields) {
    select.addEventListener("change", render);
    if (profileSelect) profileSelect.addEventListener("change", applyProfileDefaults);
    if (favoriteSelect) favoriteSelect.addEventListener("change", applyFavorite);
    if (execForm) {
      execForm.addEventListener("submit", () => {
        updateSmartWizardState(currentOperation());
      });
    }
    if (profileSelect && formValues.favorite_connection_id) profileSelect.value = String(formValues.favorite_connection_id);
    if (select && formValues.operation_id) select.value = formValues.operation_id;
    const initialOperation = currentOperation();
    if (initialOperation && initialOperation.mode) {
      activeEntity = initialOperation.mode;
    } else if (!activeEntity) {
      activeEntity = "cluster";
    }
    ensurePlaybookSelection();
    if (favoriteNameInput && formValues.favorite_name) favoriteNameInput.value = formValues.favorite_name;
    renderEntityPills(entityFilters, setActiveEntity);
    renderQuickActions();
    renderGuidedScenarios();
    render();
  }

  if (catalogSearch) {
    catalogSearch.addEventListener("input", applyCatalogSearch);
  }

  if (catalogEntityFilters) {
    renderEntityPills(catalogEntityFilters, setCatalogEntity);
    applyCatalogSearch();
  }
})();
