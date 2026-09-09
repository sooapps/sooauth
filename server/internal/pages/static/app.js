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
}

function showError(form, message) {
  let el = form.querySelector(".error");
  if (!el) {
    el = document.createElement("p");
    el.className = "error";
    form.prepend(el);
  }
  el.textContent = message;
}

function clearFeedback(form) {
  form.querySelectorAll(".error, .message").forEach((el) => el.remove());
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

const signIn = document.getElementById("sign-in-form");
if (signIn) {
  signIn.addEventListener("submit", async (e) => {
    e.preventDefault();
    clearFeedback(signIn);
    signIn.querySelector(".resend-wrap")?.remove();
    const stopLoading = setSubmitLoading(signIn, true, "Signing in…");
    const fd = new FormData(signIn);
    try {
      await postJSON("/auth/sign-in", {
        email: fd.get("email"),
        password: fd.get("password"),
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
