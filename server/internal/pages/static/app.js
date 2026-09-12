async function postJSON(path, body) {
  const res = await fetch(path, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
    credentials: "same-origin",
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    const err = new Error(data.error || "request_failed");
    err.detail = data.message || "";
    throw err;
  }
  return data;
}

const friendlyErrors = {
  invalid_credentials: "Wrong email or password.",
  email_not_verified: "Verify your email before signing in.",
  email_delivery_failed: "Email could not be sent — server SMTP may be misconfigured.",
  "email already registered": "This email is already registered. Sign in or reset your password.",
  rate_limited: "Too many attempts. Wait a minute.",
  signin_failed: "Sign-in failed. Try again.",
  request_failed: "Something went wrong. Try again.",
  passkey_failed: "Passkey sign-in failed. Use email or Google.",
};

function humanError(code) {
  return friendlyErrors[code] || "Something went wrong. Try again.";
}

function errorText(err) {
  if (err.detail) return err.detail;
  return humanError(err.message);
}

function showMessage(form, message) {
  let el = form.querySelector(".message");
  if (!el) {
    el = document.createElement("p");
    el.className = "message";
    form.prepend(el);
  }
  el.textContent = message;
  syncSplitHeight();
}

function showResendVerification(form, email) {
  let wrap = form.querySelector(".resend-wrap");
  if (wrap) return;

  wrap = document.createElement("div");
  wrap.className = "resend-wrap links";
  const btn = document.createElement("button");
  btn.type = "button";
  btn.className = "btn ghost";
  btn.textContent = "Resend verification email";
  btn.addEventListener("click", async () => {
    btn.disabled = true;
    try {
      await postJSON("/auth/resend-verification", { email });
      showMessage(form, "Verification email sent. Check inbox and spam.");
    } catch (err) {
      showError(form, errorText(err));
    } finally {
      btn.disabled = false;
    }
  });
  wrap.appendChild(btn);
  form.appendChild(wrap);
  syncSplitHeight();
}

function showError(form, message) {
  let el = form.querySelector(".error");
  if (!el) {
    el = document.createElement("p");
    el.className = "error";
    form.prepend(el);
  }
  el.textContent = message;
  syncSplitHeight();
}

function clearFeedback(form) {
  form.querySelectorAll(".error, .message").forEach((el) => el.remove());
  syncSplitHeight();
}

function setSubmitLoading(form, loading, busyLabel) {
  const btn = form.querySelector('button[type="submit"]');
  if (!btn) return () => {};
  if (!btn.dataset.defaultLabel) {
    btn.dataset.defaultLabel = btn.textContent;
  }
  btn.disabled = loading;
  btn.setAttribute("aria-busy", loading ? "true" : "false");
  if (loading) {
    btn.classList.add("is-loading");
    btn.textContent = busyLabel || "Please wait…";
  } else {
    btn.classList.remove("is-loading");
    btn.textContent = btn.dataset.defaultLabel;
  }
  return () => setSubmitLoading(form, false);
}

function afterLoginRedirect() {
  const params = new URLSearchParams(location.search);
  location.href = params.get("return_to") || "/dashboard/";
}

/* ------------------------------------------------------------------
 * Split Screen Adaptive Height & Mode Toggling
 * ------------------------------------------------------------------ */
function syncSplitHeight() {
  const container = document.getElementById("auth-split-container");
  if (!container) return;

  if (window.innerWidth <= 860) {
    container.style.minHeight = "";
    return;
  }

  const isSignUp = container.classList.contains("mode-signup");
  const activeSide = isSignUp
    ? document.getElementById("sign-up-side")
    : document.getElementById("sign-in-side");
  const infoPanel = document.getElementById("auth-panel-info");

  const activeInner = activeSide ? (activeSide.querySelector(".form-inner") || activeSide) : null;
  const infoInner = infoPanel ? (infoPanel.querySelector(".info-inner") || infoPanel) : null;

  const activeH = activeInner ? activeInner.scrollHeight + 88 : 0;
  const infoH = infoInner ? infoInner.scrollHeight + 88 : 0;

  const targetH = Math.max(activeH, infoH, 680);
  container.style.minHeight = targetH + "px";
}

function setupSplitAuth() {
  const container = document.getElementById("auth-split-container");
  if (!container) return;

  const switchBtn = document.getElementById("auth-switch-btn");
  const ctaLabel = document.getElementById("info-cta-label");
  const mobileLinks = document.querySelectorAll(".mobile-switch-link");

  function setMode(mode, pushUrl = true) {
    if (mode === "signup") {
      container.classList.remove("mode-signin");
      container.classList.add("mode-signup");
      if (switchBtn) {
        switchBtn.dataset.target = "signin";
        switchBtn.textContent = "Sign in →";
      }
      if (ctaLabel) {
        ctaLabel.textContent = "Already have an account?";
      }
      document.title = "Create account · " + (document.title.split(" · ")[1] || "sooauth");
      if (pushUrl && !location.pathname.endsWith("/sign-up")) {
        history.pushState(null, "", "/auth/sign-up" + location.search);
      }
    } else {
      container.classList.remove("mode-signup");
      container.classList.add("mode-signin");
      if (switchBtn) {
        switchBtn.dataset.target = "signup";
        switchBtn.textContent = "Create an account →";
      }
      if (ctaLabel) {
        ctaLabel.textContent = "Don't have an account yet?";
      }
      document.title = "Sign in · " + (document.title.split(" · ")[1] || "sooauth");
      if (pushUrl && !location.pathname.endsWith("/sign-in")) {
        history.pushState(null, "", "/auth/sign-in" + location.search);
      }
    }
    syncSplitHeight();
  }

  if (switchBtn) {
    switchBtn.addEventListener("click", () => {
      const isCurrentlySignin = container.classList.contains("mode-signin");
      setMode(isCurrentlySignin ? "signup" : "signin", true);
    });
  }

  mobileLinks.forEach((link) => {
    link.addEventListener("click", (e) => {
      e.preventDefault();
      const href = link.getAttribute("href") || "";
      setMode(href.includes("sign-up") ? "signup" : "signin", true);
    });
  });

  window.addEventListener("popstate", () => {
    if (location.pathname.includes("sign-up")) {
      setMode("signup", false);
    } else if (location.pathname.includes("sign-in")) {
      setMode("signin", false);
    }
  });

  window.addEventListener("resize", syncSplitHeight);
  syncSplitHeight();
}

/* ------------------------------------------------------------------
 * Theme Toggle (Dark / Light)
 * ------------------------------------------------------------------ */
function setupThemeToggle() {
  const toggleBtn = document.getElementById("theme-toggle");
  if (!toggleBtn) return;

  function getCurrentTheme() {
    return document.documentElement.getAttribute("data-theme") ||
      (window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light");
  }

  toggleBtn.addEventListener("click", () => {
    const current = getCurrentTheme();
    const next = current === "dark" ? "light" : "dark";
    document.documentElement.setAttribute("data-theme", next);
    try {
      localStorage.setItem("sooauth_theme", next);
    } catch (e) {}
  });

  // Listen for system theme changes if user hasn't explicitly set localStorage
  if (window.matchMedia) {
    window.matchMedia("(prefers-color-scheme: dark)").addEventListener("change", (e) => {
      try {
        if (!localStorage.getItem("sooauth_theme")) {
          document.documentElement.setAttribute("data-theme", e.matches ? "dark" : "light");
        }
      } catch (err) {}
    });
  }
}

// Initialize on DOM ready
document.addEventListener("DOMContentLoaded", () => {
  setupSplitAuth();
  setupThemeToggle();
});

// Direct initialization in case script runs after DOM is ready
if (document.readyState === "interactive" || document.readyState === "complete") {
  setupSplitAuth();
  setupThemeToggle();
}

/* ------------------------------------------------------------------
 * Form Submit Handlers
 * ------------------------------------------------------------------ */
const signIn = document.getElementById("sign-in-form");
if (signIn) {
  signIn.addEventListener("submit", async (e) => {
    e.preventDefault();
    clearFeedback(signIn);
    signIn.querySelector(".resend-wrap")?.remove();
    const stopLoading = setSubmitLoading(signIn, true, "Signing in…");
    const fd = new FormData(signIn);
    const rememberMeInput = signIn.querySelector('input[name="remember_me"]');
    const rememberMe = rememberMeInput ? rememberMeInput.checked : true;
    try {
      await postJSON("/auth/sign-in", {
        email: fd.get("email"),
        password: fd.get("password"),
        remember_me: rememberMe,
      });
      stopLoading();
      setSubmitLoading(signIn, true, "Redirecting…");
      afterLoginRedirect();
    } catch (err) {
      stopLoading();
      const email = fd.get("email");
      showError(signIn, errorText(err));
      if (err.message === "email_not_verified" && email) {
        showResendVerification(signIn, email);
      }
    }
  });
}

const signUp = document.getElementById("sign-up-form");
if (signUp) {
  signUp.addEventListener("submit", async (e) => {
    e.preventDefault();
    clearFeedback(signUp);
    const stopLoading = setSubmitLoading(signUp, true, "Creating account…");
    const fd = new FormData(signUp);
    try {
      await postJSON("/auth/sign-up", {
        email: fd.get("email"),
        password: fd.get("password"),
      });
      stopLoading();
      setSubmitLoading(signUp, true, "Redirecting…");
      location.href = "/auth/sign-in?message=Check+your+email+to+verify";
    } catch (err) {
      stopLoading();
      showError(signUp, errorText(err));
    }
  });
}

const forgot = document.getElementById("forgot-form");
if (forgot) {
  forgot.addEventListener("submit", async (e) => {
    e.preventDefault();
    clearFeedback(forgot);
    const stopLoading = setSubmitLoading(forgot, true, "Sending…");
    const fd = new FormData(forgot);
    const clientId = forgot.dataset.clientId || "";
    const returnTo = forgot.dataset.returnTo || "";
    try {
      const res = await postJSON("/auth/forgot-password", {
        email: fd.get("email"),
        client_id: clientId,
        return_to: returnTo,
      });
      stopLoading();
      const msg = res && res.delivery === "code"
        ? "If an account exists, we sent a 6-digit reset code to your email."
        : "If an account exists, we sent reset instructions.";
      showMessage(forgot, msg);
    } catch (err) {
      stopLoading();
      showError(forgot, errorText(err));
    }
  });
}

const reset = document.getElementById("reset-form");
if (reset) {
  reset.addEventListener("submit", async (e) => {
    e.preventDefault();
    clearFeedback(reset);
    const stopLoading = setSubmitLoading(reset, true, "Updating…");
    const fd = new FormData(reset);
    const clientId = reset.dataset.clientId || "";
    const returnTo = reset.dataset.returnTo || "";
    try {
      await postJSON("/auth/reset-password", {
        token: reset.dataset.token,
        password: fd.get("password"),
      });
      stopLoading();
      setSubmitLoading(reset, true, "Redirecting…");
      if (returnTo) {
        location.href = returnTo;
      } else {
        const q = new URLSearchParams();
        q.set("message", "Password updated");
        if (clientId) q.set("client_id", clientId);
        location.href = "/auth/sign-in?" + q.toString();
      }
    } catch (err) {
      stopLoading();
      showError(reset, errorText(err));
    }
  });
}

const passkeyBtn = document.getElementById("passkey-sign-in");
if (passkeyBtn) {
  passkeyBtn.addEventListener("click", async () => {
    const form = document.getElementById("sign-in-form");
    const emailInput = form?.querySelector('input[name="email"]');
    const email = emailInput?.value?.trim();
    if (!email) {
      showError(form, "Enter your email first.");
      emailInput?.focus();
      return;
    }
    try {
      const begin = await postJSON("/auth/passkey/sign-in/begin", { email });
      const cred = await navigator.credentials.get({ publicKey: begin.publicKey });
      const finish = await fetch("/auth/passkey/sign-in/finish", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "same-origin",
        body: JSON.stringify({ state: begin.state, credential: cred }),
      });
      if (!finish.ok) throw new Error("passkey_failed");
      afterLoginRedirect();
    } catch (err) {
      showError(form, errorText(err));
    }
  });
}
