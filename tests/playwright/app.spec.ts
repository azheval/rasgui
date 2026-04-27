import { expect, test } from "@playwright/test";

test("login page supports language switching", async ({ page }) => {
  await page.goto("/set-lang?lang=en&next=/login");
  await expect(page.getByRole("button", { name: "Sign In" })).toBeVisible();

  await page.goto("/set-lang?lang=ru&next=/login");
  await expect(page.getByRole("button").first()).toBeVisible();

  await page.goto("/set-lang?lang=be&next=/login");
  await expect(page.getByRole("button").first()).toBeVisible();
});

test("admin can sign in and open main sections", async ({ page }) => {
  await page.goto("/set-lang?lang=en&next=/login");
  await page.getByLabel("Username").fill("admin");
  await page.getByLabel("Password").fill("admin123");
  await page.getByRole("button", { name: "Sign In" }).click();

  await expect(page.getByRole("heading", { name: "rasgui" })).toBeVisible();
  await expect(page.getByText("Default remote RAS")).toBeVisible();

  await page.getByRole("link", { name: "Catalog" }).click();
  await expect(page.getByText("Command catalog")).toBeVisible();

  await page.getByRole("link", { name: "Execute" }).click();
  await expect(page.getByText("Execute operation")).toBeVisible();
});

test("admin can create a remote RAS connection profile", async ({ page }) => {
  await page.goto("/set-lang?lang=en&next=/login");
  await page.getByLabel("Username").fill("admin");
  await page.getByLabel("Password").fill("admin123");
  await page.getByRole("button", { name: "Sign In" }).click();

  await page.getByRole("link", { name: "Connections" }).click();
  const connectionCard = page.locator(".card").filter({ hasText: "Create connection profile" }).first();
  await connectionCard.getByLabel("Name").fill("demo-ras");
  await connectionCard.getByLabel("Host").fill("demo-host");
  await connectionCard.getByLabel("Port").fill("1645");
  await connectionCard.getByLabel("Description").fill("Demo remote server");
  await connectionCard.getByRole("button", { name: "Create connection profile" }).click();

  const savedConnection = page.locator('.compact-form input[name="name"][value="demo-ras"]').first();
  await expect(savedConnection).toBeVisible();
  await expect(page.locator('.compact-form input[name="host"][value="demo-host"]').first()).toBeVisible();
});

test("admin can create a dedicated toolchain profile and bind it to a connection", async ({ page }) => {
  const suffix = Date.now().toString();
  const toolchainName = `1c-test-${suffix}`;
  const connectionName = `conn-${suffix}`;

  await page.goto("/set-lang?lang=en&next=/login");
  await page.getByLabel("Username").fill("admin");
  await page.getByLabel("Password").fill("admin123");
  await page.getByRole("button", { name: "Sign In" }).click();

  await page.getByRole("link", { name: "Connections" }).click();

  const toolchainCard = page.locator(".card").filter({ hasText: "Create toolchain profile" }).first();
  await toolchainCard.getByLabel("Name").fill(toolchainName);
  await toolchainCard.getByLabel("Version").fill("8.3.99.1");
  await toolchainCard.getByLabel("RAC path").fill("C:\\tools\\1c\\8.3.99.1\\rac.exe");
  await toolchainCard.getByLabel("RAS path").fill("C:\\tools\\1c\\8.3.99.1\\ras.exe");
  await toolchainCard.getByRole("button", { name: "Create toolchain profile" }).click();

  const toolchainRow = page.locator("table tbody tr").filter({ hasText: toolchainName }).first();
  await expect(toolchainRow).toContainText("8.3.99.1");

  const connectionCard = page.locator(".card").filter({ hasText: "Create connection profile" }).first();
  await connectionCard.getByLabel("Name").fill(connectionName);
  await connectionCard.getByLabel("Host").fill("demo-multi-host");
  await connectionCard.getByLabel("Port").fill("1745");
  await connectionCard.getByLabel("Toolchain").selectOption({ label: `${toolchainName} (8.3.99.1)` });
  await connectionCard.getByRole("button", { name: "Create connection profile" }).click();

  const savedConnectionForm = page.locator(".compact-form").filter({ has: page.locator(`input[name="name"][value="${connectionName}"]`) }).first();
  await expect(savedConnectionForm).toBeVisible();
  await expect(savedConnectionForm.locator('select[name="toolchain_id"]')).toHaveValue(/.+/);
});

test("favorites are empty on a fresh execute screen until the user saves one", async ({ page }) => {
  await page.goto("/set-lang?lang=en&next=/login");
  await page.getByLabel("Username").fill("admin");
  await page.getByLabel("Password").fill("admin123");
  await page.getByRole("button", { name: "Sign In" }).click();

  await page.getByRole("link", { name: "Execute" }).click();
  await page.getByLabel("Technical command").selectOption("rac.cluster.list");

  await expect(page.locator("#favorite_select")).toHaveValue("");
  await expect(page.locator("#favorite_select option")).toHaveCount(1);
  await expect(page.locator("#favorite_note")).toContainText("No favorites saved yet");
});

test("admin can save and reuse favorite command", async ({ page }) => {
  const suffix = Date.now().toString();
  const clusterFavorite = `Cluster list favorite ${suffix}`;
  const sessionFavorite = `Session list favorite ${suffix}`;

  await page.goto("/set-lang?lang=en&next=/login");
  await page.getByLabel("Username").fill("admin");
  await page.getByLabel("Password").fill("admin123");
  await page.getByRole("button", { name: "Sign In" }).click();

  await page.getByRole("link", { name: "Execute" }).click();
  await page.getByLabel("Technical command").selectOption("rac.cluster.list");
  await page.getByLabel("Favorite name").fill(clusterFavorite);
  await page.getByRole("button", { name: "Save favorite" }).click();

  await expect(page.getByLabel("Technical command")).toHaveValue("rac.cluster.list");
  await expect(page.getByLabel("Favorite name")).toHaveValue("");
  await expect(page.locator("#favorite_select")).toContainText(clusterFavorite);

  await page.getByLabel("Technical command").selectOption("rac.session.list");
  await expect(page.locator("#favorite_select")).not.toContainText(clusterFavorite);

  await page.getByLabel("Favorite name").fill(sessionFavorite);
  await page.getByRole("button", { name: "Save favorite" }).click();
  await expect(page.locator("#favorite_select")).toContainText(sessionFavorite);
  await expect(page.locator("#favorite_select")).not.toContainText(clusterFavorite);

  await page.getByLabel("Technical command").selectOption("rac.cluster.list");
  await page.locator("#favorite_select").selectOption({ label: clusterFavorite });
  await expect(page.getByLabel("Technical command")).toHaveValue("rac.cluster.list");
});

test("guided scenarios help select common operations", async ({ page }) => {
  await page.goto("/set-lang?lang=en&next=/login");
  await page.getByLabel("Username").fill("admin");
  await page.getByLabel("Password").fill("admin123");
  await page.getByRole("button", { name: "Sign In" }).click();

  await page.getByRole("link", { name: "Execute" }).click();
  await page.locator("#entity_filters").getByRole("button", { name: "Session", exact: true }).click();
  await page.locator("#guided_scenarios .wizard-tile").nth(2).click();

  await expect(page.getByLabel("Technical command")).toHaveValue("rac.session.terminate");
});

test("human-friendly action cards select the technical command automatically", async ({ page }) => {
  await page.goto("/set-lang?lang=en&next=/login");
  await page.getByLabel("Username").fill("admin");
  await page.getByLabel("Password").fill("admin123");
  await page.getByRole("button", { name: "Sign In" }).click();

  await page.getByRole("link", { name: "Execute" }).click();
  await page.locator("#entity_filters").getByRole("button", { name: "Cluster", exact: true }).click();
  await page.locator("#playbook_list").getByRole("button", { name: /Access/i }).click();
  await page.locator("#human_action_list").getByRole("button", { name: /Cluster administrators/i }).click();

  await expect(page.getByLabel("Technical command")).toHaveValue("rac.cluster.admin.list");
  await expect(page.locator("#journey_summary")).toContainText("Cluster administrators");
});

test("entity wizard tracks narrow actions to the selected work scenario", async ({ page }) => {
  await page.goto("/set-lang?lang=en&next=/login");
  await page.getByLabel("Username").fill("admin");
  await page.getByLabel("Password").fill("admin123");
  await page.getByRole("button", { name: "Sign In" }).click();

  await page.getByRole("link", { name: "Execute" }).click();
  await page.locator("#entity_filters").getByRole("button", { name: "Cluster", exact: true }).click();
  await page.locator("#playbook_list").getByRole("button", { name: /Access/i }).click();

  await expect(page.locator("#journey_summary")).toContainText("Access");
  await expect(page.locator("#human_action_list")).toContainText("Cluster administrators");
  await expect(page.locator("#human_action_list")).not.toContainText("Cluster update");
});

test("infobase maintenance wizard assembles lock parameters automatically", async ({ page }) => {
  await page.goto("/set-lang?lang=en&next=/login");
  await page.getByLabel("Username").fill("admin");
  await page.getByLabel("Password").fill("admin123");
  await page.getByRole("button", { name: "Sign In" }).click();

  await page.getByRole("link", { name: "Execute" }).click();
  await page.getByLabel("Technical command").selectOption("rac.infobase.update");

  await expect(page.locator("#smart_wizard_host")).toContainText("Maintenance and lock wizard");
  await page.getByRole("button", { name: "Lock for maintenance" }).click();
  await page.getByLabel("User message").fill("Maintenance window");
  await page.getByLabel("Permission code").fill("letmein");

  await expect(page.locator("#smart_wizard_preview")).toContainText("--sessions-deny=on");
  await expect(page.locator("#smart_wizard_preview")).toContainText("--scheduled-jobs-deny=on");
  await expect(page.locator("#smart_wizard_preview")).toContainText("--denied-message=Maintenance window");
  await expect(page.locator("#smart_wizard_preview")).toContainText("--permission-code=letmein");
  await expect(page.locator("#wizard_args_json")).toHaveValue(/--sessions-deny=on/);
  await expect(page.locator("#wizard_args_json")).toHaveValue(/--permission-code=letmein/);
});

test("infobase create wizard applies scenario presets", async ({ page }) => {
  await page.goto("/set-lang?lang=en&next=/login");
  await page.getByLabel("Username").fill("admin");
  await page.getByLabel("Password").fill("admin123");
  await page.getByRole("button", { name: "Sign In" }).click();

  await page.getByRole("link", { name: "Execute" }).click();
  await page.getByLabel("Technical command").selectOption("rac.infobase.create");

  await expect(page.locator("#smart_wizard_host")).toContainText("Infobase creation wizard");
  await page.getByRole("button", { name: "New PostgreSQL infobase" }).click();
  await expect(page.getByLabel("DBMS")).toHaveValue("PostgreSQL");
  await expect(page.getByLabel("Create database automatically")).toBeChecked();
});

test("cluster admin wizard switches authentication scenario presets", async ({ page }) => {
  await page.goto("/set-lang?lang=en&next=/login");
  await page.getByLabel("Username").fill("admin");
  await page.getByLabel("Password").fill("admin123");
  await page.getByRole("button", { name: "Sign In" }).click();

  await page.getByRole("link", { name: "Execute" }).click();
  await page.getByLabel("Technical command").selectOption("rac.cluster.admin.register");

  await expect(page.locator("#smart_wizard_host")).toContainText("Cluster admin registration wizard");
  await page.getByRole("button", { name: "OS-backed administrator" }).click();
  await expect(page.getByLabel("Authentication mode")).toHaveValue("pwd,os");
});

test("result context reuses identifiers from the latest command output", async ({ page }) => {
  await page.goto("/set-lang?lang=en&next=/login");
  await page.getByLabel("Username").fill("admin");
  await page.getByLabel("Password").fill("admin123");
  await page.getByRole("button", { name: "Sign In" }).click();

  await page.getByRole("link", { name: "Execute" }).click();
  await page.getByLabel("Technical command").selectOption("rac.cluster.list");
  await page.getByRole("button", { name: "Execute" }).click();
  await page.getByLabel("Technical command").selectOption("rac.cluster.info");
  await page.locator(".result-context-chip").filter({ hasText: "cluster" }).first().click();

  await expect(page.locator('input[name="cluster"]').first()).not.toHaveValue("");
});

test("cluster update wizard assembles tuning profile", async ({ page }) => {
  await page.goto("/set-lang?lang=en&next=/login");
  await page.getByLabel("Username").fill("admin");
  await page.getByLabel("Password").fill("admin123");
  await page.getByRole("button", { name: "Sign In" }).click();

  await page.getByRole("link", { name: "Execute" }).click();
  await page.getByLabel("Technical command").selectOption("rac.cluster.update");
  await expect(page.locator("#smart_wizard_host")).toContainText("Cluster tuning wizard");
  await page.getByRole("button", { name: "Memory-priority" }).click();
  await expect(page.locator("#smart_wizard_preview")).toContainText("--load-balancing-mode=memory");
});

test("session and connection wizards expose guided fields", async ({ page }) => {
  await page.goto("/set-lang?lang=en&next=/login");
  await page.getByLabel("Username").fill("admin");
  await page.getByLabel("Password").fill("admin123");
  await page.getByRole("button", { name: "Sign In" }).click();

  await page.getByRole("link", { name: "Execute" }).click();
  await page.getByLabel("Technical command").selectOption("rac.session.terminate");
  await expect(page.locator("#smart_wizard_host")).toContainText("Session termination wizard");
  await expect(page.getByLabel("Session UUID")).toBeVisible();

  await page.getByLabel("Technical command").selectOption("rac.connection.disconnect");
  await expect(page.locator("#smart_wizard_host")).toContainText("Connection disconnect wizard");
  await expect(page.getByLabel("Worker process UUID")).toBeVisible();
});

test("operator flows guide the user through a maintenance chain", async ({ page }) => {
  await page.goto("/set-lang?lang=en&next=/login");
  await page.getByLabel("Username").fill("admin");
  await page.getByLabel("Password").fill("admin123");
  await page.getByRole("button", { name: "Sign In" }).click();

  await page.getByRole("link", { name: "Execute" }).click();
  await expect(page.locator("#operator_flow_host")).toContainText("Infobase maintenance");
  await page.getByRole("button", { name: "Infobase maintenance" }).click();
  await page.getByRole("button", { name: "Start flow" }).click();
  await expect(page.getByLabel("Technical command")).toHaveValue("rac.infobase.update");
  await page.getByRole("button", { name: "Next step" }).click();
  await expect(page.getByLabel("Technical command")).toHaveValue("rac.session.list");
});

test("operator flows hide required-inaccessible chains and trim optional steps by role", async ({ page, context }) => {
  const suffix = Date.now().toString();
  const roleName = `ib-maintainer-${suffix}`;
  const username = `ib_ops_${suffix}`;
  const password = "viewer123";

  await page.goto("/set-lang?lang=en&next=/login");
  await page.getByLabel("Username").fill("admin");
  await page.getByLabel("Password").fill("admin123");
  await page.getByRole("button", { name: "Sign In" }).click();

  await page.getByRole("link", { name: "Roles" }).click();
  await page.getByLabel("Name").fill(roleName);
  await page.getByLabel("Description").fill("Infobase maintenance lock only");
  await page.getByRole("button", { name: "Create role" }).click();

  const roleBox = page.locator(".role-box").filter({ hasText: roleName }).first();
  const scopeCard = roleBox.locator(".scope-set-card").last();
  const infobaseSection = scopeCard.locator(".matrix-section").filter({ hasText: "infobase" }).first();
  await infobaseSection.locator('input[type="checkbox"][value="rac.infobase.update"]').check();
  await scopeCard.locator('textarea[name="cluster_scope"]').fill("*");
  await scopeCard.locator('textarea[name="infobase_scope"]').fill("*");
  await scopeCard.getByRole("button", { name: "Add scope set" }).click();

  await page.getByRole("link", { name: "Users" }).click();
  await page.getByLabel("Username").fill(username);
  await page.getByLabel("Password").fill(password);
  await page.getByLabel("Role").selectOption({ label: roleName });
  await page.getByRole("button", { name: "Create user" }).click();

  await context.clearCookies();

  await page.goto("/set-lang?lang=en&next=/login");
  await page.getByLabel("Username").fill(username);
  await page.getByLabel("Password").fill(password);
  await page.getByRole("button", { name: "Sign In" }).click();
  await page.getByRole("link", { name: "Execute" }).click();

  await expect(page.locator("#operator_flow_host")).toContainText("Infobase maintenance");
  await expect(page.locator("#operator_flow_host")).toContainText("Some steps are hidden by role");
  await expect(page.locator("#operator_flow_host")).not.toContainText("Session cleanup");
});

test("limited user only sees operations allowed by role", async ({ page, context }) => {
  const suffix = Date.now().toString();
  const roleName = `cluster-list-only-${suffix}`;
  const username = `viewer_${suffix}`;
  const password = "viewer123";

  await page.goto("/set-lang?lang=en&next=/login");
  await page.getByLabel("Username").fill("admin");
  await page.getByLabel("Password").fill("admin123");
  await page.getByRole("button", { name: "Sign In" }).click();

  await page.getByRole("link", { name: "Roles" }).click();
  await page.getByLabel("Name").fill(roleName);
  await page.getByLabel("Description").fill("Limited cluster list access");
  await page.getByRole("button", { name: "Create role" }).click();

  const limitedRoleBox = page.locator(".role-box").filter({ hasText: roleName }).first();
  await limitedRoleBox.locator('input[type="checkbox"][value="rac.cluster.list"]').check();
  await limitedRoleBox.getByRole("button", { name: "Add scope set" }).click();

  await page.getByRole("link", { name: "Users" }).click();
  await page.getByLabel("Username").fill(username);
  await page.getByLabel("Password").fill(password);
  await page.getByLabel("Role").selectOption({ label: roleName });
  await page.getByRole("button", { name: "Create user" }).click();

  await context.clearCookies();

  await page.goto("/set-lang?lang=en&next=/login");
  await page.getByLabel("Username").fill(username);
  await page.getByLabel("Password").fill(password);
  await page.getByRole("button", { name: "Sign In" }).click();

  await expect(page.getByRole("link", { name: "Users" })).not.toBeVisible();
  await expect(page.getByRole("link", { name: "Roles" })).not.toBeVisible();
  await expect(page.getByRole("link", { name: "Connections" })).not.toBeVisible();
  await expect(page.getByRole("link", { name: "Audit" })).not.toBeVisible();

  await page.getByRole("link", { name: "Catalog" }).click();
  await expect(page.getByText("rac.cluster.list")).toBeVisible();
  await expect(page.getByText("rac.cluster.info")).not.toBeVisible();
  await expect(page.getByText("rac.session.list")).not.toBeVisible();

  await page.getByRole("link", { name: "Execute" }).click();
  await expect(page.locator("#entity_filters").getByRole("button", { name: "Cluster", exact: true })).toBeVisible();
  await expect(page.locator("#entity_filters").getByRole("button", { name: "Session", exact: true })).not.toBeVisible();
  await expect(page.locator("#guided_scenarios")).toContainText("Cluster list");
  await expect(page.locator("#guided_scenarios")).not.toContainText("Cluster administrators");
  await expect(page.locator("#human_action_list")).toContainText("Cluster list");
  await expect(page.locator("#human_action_list")).not.toContainText("Cluster update");
  await expect(page.getByLabel("Technical command").locator("option")).toHaveCount(2);

  const usersResponse = await page.goto("/users");
  expect(usersResponse?.status()).toBe(403);
  const rolesResponse = await page.goto("/roles");
  expect(rolesResponse?.status()).toBe(403);
  const connectionsResponse = await page.goto("/connections");
  expect(connectionsResponse?.status()).toBe(403);
  const auditResponse = await page.goto("/audit");
  expect(auditResponse?.status()).toBe(403);
});

test("admin can manage users and roles with access matrix and delete them", async ({ page }) => {
  const suffix = Date.now().toString();
  const roleName = `matrix-role-${suffix}`;
  const username = `matrix_user_${suffix}`;

  await page.goto("/set-lang?lang=en&next=/login");
  await page.getByLabel("Username").fill("admin");
  await page.getByLabel("Password").fill("admin123");
  await page.getByRole("button", { name: "Sign In" }).click();

  await page.getByRole("link", { name: "Roles" }).click();
  await page.getByLabel("Name").fill(roleName);
  await page.getByLabel("Description").fill("Matrix-managed role");
  await page.getByRole("button", { name: "Create role" }).click();

  const roleBox = page.locator(".role-box").filter({ hasText: roleName }).first();
  const newScopeCard = roleBox.locator(".scope-set-card").last();
  const clusterSection = newScopeCard.locator(".matrix-section").filter({ hasText: "cluster" }).first();
  await clusterSection.getByRole("button", { name: "View" }).click();
  await newScopeCard.locator('textarea[name="cluster_scope"]').fill("cluster-alpha");
  await newScopeCard.locator('textarea[name="infobase_scope"]').fill("*");
  await newScopeCard.getByRole("button", { name: "Add scope set" }).click();

  await expect(roleBox).toContainText("rac.cluster.list");
  await expect(roleBox).toContainText("rac.cluster.info");
  await expect(roleBox).toContainText("cluster-alpha");

  const secondScopeCard = roleBox.locator(".scope-set-card").last();
  const secondClusterSection = secondScopeCard.locator(".matrix-section").filter({ hasText: "cluster" }).first();
  await secondClusterSection.getByRole("button", { name: "Change" }).click();
  await secondScopeCard.locator('textarea[name="cluster_scope"]').fill("cluster-beta");
  await secondScopeCard.locator('textarea[name="infobase_scope"]').fill("*");
  await secondScopeCard.getByRole("button", { name: "Add scope set" }).click();

  await expect(roleBox).toContainText("cluster-beta");
  await expect(roleBox.getByRole("button", { name: "Delete scope set" })).toHaveCount(2);

  await page.getByRole("link", { name: "Users" }).click();
  await page.getByLabel("Username").fill(username);
  await page.getByLabel("Password").fill("user12345");
  await page.getByLabel("Role").selectOption({ label: roleName });
  await page.getByRole("button", { name: "Create user" }).click();

  const userRow = page.locator("tbody tr").filter({ hasText: username }).first();
  await expect(userRow).toContainText(roleName);
  await userRow.getByRole("button", { name: "Delete user" }).click();
  await expect(page.locator("tbody tr").filter({ hasText: username })).toHaveCount(0);

  await page.getByRole("link", { name: "Roles" }).click();
  await page.locator(".role-box").filter({ hasText: roleName }).first().getByRole("button", { name: "Delete role" }).click();
  await expect(page.locator(".role-box").filter({ hasText: roleName })).toHaveCount(0);
});
