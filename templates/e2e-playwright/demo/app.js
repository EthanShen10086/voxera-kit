const DEMO_USER = { email: "demo@voxera.test", password: "demo-pass" };

function readSession() {
  try {
    return JSON.parse(localStorage.getItem("demo-session") ?? "null");
  } catch {
    return null;
  }
}

function writeSession(session) {
  localStorage.setItem("demo-session", JSON.stringify(session));
}

function clearSession() {
  localStorage.removeItem("demo-session");
}

function initLoginPage() {
  const form = document.getElementById("login-form");
  if (!form) return;

  form.addEventListener("submit", (event) => {
    event.preventDefault();
    const data = new FormData(form);
    const email = String(data.get("email") ?? "");
    const password = String(data.get("password") ?? "");
    const error = document.querySelector("[data-testid='error']");

    if (email === DEMO_USER.email && password === DEMO_USER.password) {
      writeSession({ email, token: "demo-token" });
      window.location.href = "./dashboard.html";
      return;
    }

    if (error) {
      error.hidden = false;
      error.textContent = "Invalid email or password";
    }
  });
}

function initDashboardPage() {
  const welcome = document.querySelector("[data-testid='welcome']");
  if (!welcome) return;

  const session = readSession();
  if (!session?.token) {
    window.location.href = "./login.html";
    return;
  }

  const email = document.querySelector("[data-testid='user-email']");
  if (email) email.textContent = session.email;

  document.querySelector("[data-testid='logout']")?.addEventListener("click", () => {
    clearSession();
    window.location.href = "./login.html";
  });
}

if (document.getElementById("login-form")) {
  initLoginPage();
} else if (document.querySelector("[data-testid='welcome']")) {
  initDashboardPage();
}
