/* ------------------------------------------------------------------
 * Hosted sign-in for a project's end users (pages opened with a
 * client_id, e.g. from /oauth/authorize). Reads the project's settings
 * from /v1/widget/config and supports identifiers, password / one-time
 * code / both, two-step verification and extra sign-up fields. Every
 * successful request sends hosted: true so the server keeps the session
 * in the project cookie, separate from the dashboard session.
 * ------------------------------------------------------------------ */
(function () {
  const clientId = window.__sooauthProject;
  const signInForm = document.getElementById("sign-in-form");
  const signUpForm = document.getElementById("sign-up-form");
  if (!clientId || (!signInForm && !signUpForm)) return;

  let cfg = null;
  let preferCode = false;
  let pendingIdentifier = "";

  const esc = (s) => String(s == null ? "" : s).replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));

  async function call(path, body) {
    const res = await fetch(path, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
      credentials: "same-origin",
    });
    const data = await res.json().catch(() => ({}));
    return { ok: res.ok, status: res.status, data };
  }

  function message(form, text, isError) {
    clearFeedback(form);
    if (!text) return;
    if (isError) {
      showError(form, text);
      return;
    }
    const p = document.createElement("p");
    p.className = "message";
    p.textContent = text;
    form.prepend(p);
  }

  function errorFor(data) {
    switch (data.error) {
      case "invalid_code":
        return t("js.p.invalid_code", "That code is not right. Try again.");
      case "too_many_attempts":
        return t("js.p.too_many", "Too many wrong codes. Request a new code.");
      case "password_login_disabled":
        return t("js.p.code_only", "This app signs in with one-time codes.");
      case "otp_disabled":
        return t("js.p.otp_disabled", "Sign-in codes are not enabled for this app.");
      case "phone_otp_unavailable":
        return t("js.p.phone_unavailable", "Codes to phone numbers are not set up for this app.");
      case "invalid_mfa_code":
        return t("js.p.invalid_code", "That code is not right. Try again.");
      default:
        return data.message || humanError(data.error || "request_failed");
    }
  }

  function identifierLabel() {
    const ids = (cfg && cfg.auth_config && cfg.auth_config.allowed_identifiers) || ["email"];
    const names = {
      email: t("js.p.id_email", "Email"),
      username: t("js.p.id_username", "Username"),
      phone: t("js.p.id_phone", "Phone number"),
    };
    return ids.map((id) => names[id] || id).join(" / ");
  }

  const mode = () => (cfg && cfg.auth_config && cfg.auth_config.primary_auth_mode) || "password";
  const usingCode = () => mode() === "otp" || (mode() === "both" && preferCode);

  /* ---------------- sign in ---------------- */

  function renderSignIn() {
    if (!signInForm) return;
    const idInput = signInForm.querySelector('input[name="email"]');
    const pwLabel = signInForm.querySelector('input[name="password"]')?.closest("label");
    const remember = signInForm.querySelector(".checkbox-row");
    const submit = signInForm.querySelector('button[type="submit"]');
    const idLabel = idInput?.closest("label");
    if (idLabel && idLabel.firstChild && idLabel.firstChild.nodeType === 3) idLabel.firstChild.textContent = identifierLabel();
    if (idInput) idInput.placeholder = identifierLabel();

    const code = usingCode();
    if (pwLabel) {
      pwLabel.hidden = code;
      pwLabel.querySelector("input").required = !code;
    }
    if (remember) remember.hidden = code;
    if (submit) submit.textContent = code ? t("js.p.send_code", "Send code") : t("js.p.sign_in", "Sign in");
    const forgot = document.querySelector("#sign-in-side .links a[href*='forgot-password']");
    if (forgot) forgot.hidden = mode() === "otp";

    let toggle = document.getElementById("project-mode-toggle");
    if (mode() === "both") {
      if (!toggle) {
        toggle = document.createElement("button");
        toggle.type = "button";
        toggle.id = "project-mode-toggle";
        toggle.className = "btn ghost";
        toggle.addEventListener("click", () => {
          preferCode = !preferCode;
          clearFeedback(signInForm);
          renderSignIn();
        });
        signInForm.after(toggle);
      }
      toggle.textContent = code ? t("js.p.use_password", "Use my password instead") : t("js.p.use_code", "Use a one-time code instead");
    }
  }

  function codeStep(form, opts) {
    // Replaces the form body with a 6-digit code input.
    form.dataset.step = opts.step;
    form.innerHTML =
      (opts.lead ? `<p class="message">${esc(opts.lead)}</p>` : "") +
      (opts.extra || "") +
      `<label>${esc(t("js.p.code", "Verification code"))}
        <input name="code" inputmode="numeric" autocomplete="one-time-code" maxlength="10" required placeholder="123456" />
      </label>
      <button type="submit" class="btn primary">${esc(opts.button || t("js.p.verify", "Verify"))}</button>` +
      (opts.links || "");
    form.querySelector('input[name="code"]').focus();
    syncSplitHeight();
  }

  async function sendCode(identifier) {
    const res = await call("/auth/otp/send", { client_id: clientId, identifier });
    if (!res.ok) {
      message(signInForm, errorFor(res.data), true);
      return false;
    }
    pendingIdentifier = identifier;
    codeStep(signInForm, {
      step: "otp",
      lead: t("js.p.code_sent", "We sent a 6-digit code to {id}.").replace("{id}", identifier),
      button: t("js.p.sign_in", "Sign in"),
      links: `<button type="button" class="btn ghost" id="project-resend">${esc(t("js.p.resend", "Send a new code"))}</button>`,
    });
    document.getElementById("project-resend")?.addEventListener("click", async () => {
      const again = await call("/auth/otp/send", { client_id: clientId, identifier: pendingIdentifier });
      message(signInForm, again.ok ? t("js.p.resent", "A new code is on its way.") : errorFor(again.data), !again.ok);
    });
    return true;
  }

  async function startMFA(data) {
    const state = { challenge: data.challenge, method: data.method || "totp", methods: data.methods || [] };
    let extra = "";
    let lead = t("js.p.mfa_totp", "Enter the 6-digit code from your authenticator app, or a backup code.");
    if (state.method === "email_otp") lead = t("js.p.mfa_email", "We emailed you a 6-digit code to finish signing in.");
    if (state.method === "totp_setup") {
      lead = t("js.p.mfa_setup", "This app requires two-step verification. Scan the QR code with an authenticator app, then enter the code.");
      const setup = await call("/auth/mfa/setup", { challenge: state.challenge });
      if (setup.ok) {
        extra = `<img src="${esc(setup.data.qr_code || "")}" alt="QR" width="180" height="180" style="display:block;margin:0 auto" />
          <p class="passkey-hint" style="text-align:center;word-break:break-all">${esc(t("js.p.mfa_key", "Or enter this key:"))} <code>${esc(setup.data.secret)}</code></p>`;
      }
    }
    const canEmail = state.method === "totp" && state.methods.indexOf("email_otp") !== -1;
    codeStep(signInForm, {
      step: "mfa",
      lead,
      extra,
      links:
        (canEmail ? `<button type="button" class="btn ghost" id="project-mfa-email">${esc(t("js.p.mfa_use_email", "Email me a code instead"))}</button>` : "") +
        (state.method === "email_otp" ? `<button type="button" class="btn ghost" id="project-mfa-email">${esc(t("js.p.resend", "Send a new code"))}</button>` : ""),
    });
    signInForm._mfa = state;
    document.getElementById("project-mfa-email")?.addEventListener("click", async () => {
      const res = await call("/auth/mfa/resend", { challenge: state.challenge });
      if (res.ok) state.method = "email_otp";
      message(signInForm, res.ok ? t("js.p.mfa_email", "We emailed you a 6-digit code to finish signing in.") : errorFor(res.data), !res.ok);
    });
  }

  function showBackupCodes(codes) {
    const side = signInForm.closest(".form-inner") || signInForm.parentElement;
    signInForm.innerHTML = `<p class="message">${esc(t("js.p.backup_title", "Save these backup codes. Each one works once if you lose your phone."))}</p>
      <pre style="padding:12px;border:1px solid var(--line, #ddd);border-radius:8px;columns:2;line-height:1.8">${codes.map(esc).join("\n")}</pre>
      <button type="button" class="btn primary" id="project-continue">${esc(t("js.p.continue", "Continue"))}</button>`;
    document.getElementById("project-continue").addEventListener("click", afterLoginRedirect);
    side && syncSplitHeight();
  }

  signInForm?.addEventListener("submit", async (e) => {
    e.preventDefault();
    if (!cfg) return;
    const step = signInForm.dataset.step || "";
    const fd = new FormData(signInForm);
    const stop = setSubmitLoading(signInForm, true, t("js.please_wait", "Please wait…"));
    try {
      if (step === "otp") {
        const res = await call("/auth/otp/verify", { client_id: clientId, identifier: pendingIdentifier, code: String(fd.get("code") || ""), hosted: true });
        if (res.ok) return afterLoginRedirect();
        if (res.data.error === "registration_required") {
          // New user with required sign-up fields: finish on the sign-up form.
          location.href = "/auth/sign-up" + location.search + (location.search ? "&" : "?") + "prefill=" + encodeURIComponent(pendingIdentifier);
          return;
        }
        return message(signInForm, errorFor(res.data), true);
      }
      if (step === "mfa") {
        const st = signInForm._mfa;
        const res = await call("/auth/mfa/verify", { challenge: st.challenge, code: String(fd.get("code") || ""), method: st.method, hosted: true });
        if (!res.ok) return message(signInForm, errorFor(res.data), true);
        if (res.data.backup_codes && res.data.backup_codes.length) return showBackupCodes(res.data.backup_codes);
        return afterLoginRedirect();
      }
      const identifier = String(fd.get("email") || "").trim();
      if (usingCode()) {
        await sendCode(identifier);
        return;
      }
      const remember = signInForm.querySelector('input[name="remember_me"]');
      const res = await call("/auth/sign-in", {
        identifier,
        password: String(fd.get("password") || ""),
        client_id: clientId,
        remember_me: remember ? remember.checked : true,
        hosted: true,
      });
      if (res.ok) return afterLoginRedirect();
      if (res.data.error === "mfa_required") return startMFA(res.data);
      message(signInForm, errorFor(res.data), true);
      if (res.data.error === "email_not_verified" && identifier) showResendVerification(signInForm, identifier);
    } finally {
      stop();
    }
  });

  /* ---------------- sign up ---------------- */

  function fieldHTML(f) {
    const req = f.required ? " required" : "";
    const label = esc(f.label) + (f.required ? " *" : "");
    const name = `meta_${esc(f.id)}`;
    const help = f.description ? `<small class="passkey-hint">${esc(f.description)}</small>` : "";
    if (f.type === "checkbox") {
      return `<label class="checkbox-row"><input type="checkbox" name="${name}"${req} /><span>${label}</span></label>${help}`;
    }
    if (f.type === "select") {
      const opts = (f.options || []).map((o) => `<option value="${esc(o.value)}">${esc(o.label)}</option>`).join("");
      return `<label>${label}<select name="${name}"${req}${f.options_url ? ` data-options-url="${esc(f.options_url)}"` : ""}><option value="">${esc(f.placeholder || "—")}</option>${opts}</select>${help}</label>`;
    }
    if (f.type === "textarea") {
      return `<label>${label}<textarea name="${name}" placeholder="${esc(f.placeholder || "")}"${req}></textarea>${help}</label>`;
    }
    const type = f.type === "number" ? "number" : "text";
    return `<label>${label}<input name="${name}" type="${type}" placeholder="${esc(f.placeholder || "")}"${req} />${help}</label>`;
  }

  function renderSignUp() {
    if (!signUpForm) return;
    const ac = cfg.auth_config || {};
    const ids = ac.allowed_identifiers || ["email"];
    const emailLabel = signUpForm.querySelector('input[name="email"]')?.closest("label");
    const pwLabel = signUpForm.querySelector('input[name="password"]')?.closest("label");
    const submit = signUpForm.querySelector('button[type="submit"]');
    const params = new URLSearchParams(location.search);
    const prefill = params.get("prefill") || "";

    if (emailLabel) {
      emailLabel.hidden = ids.indexOf("email") === -1;
      emailLabel.querySelector("input").required = ids.length === 1 && ids[0] === "email";
      if (prefill.indexOf("@") !== -1) emailLabel.querySelector("input").value = prefill;
    }
    const extraIds = ids.filter((id) => id !== "email").map((id) => {
      const only = ids.length === 1;
      const val = prefill && prefill.indexOf("@") === -1 ? esc(prefill) : "";
      return id === "username"
        ? `<label>${esc(t("js.p.id_username", "Username"))}${only ? " *" : ""}<input name="username" autocomplete="username"${only ? " required" : ""} value="${val}" /></label>`
        : `<label>${esc(t("js.p.id_phone", "Phone number"))}${only ? " *" : ""}<input name="phone" type="tel" placeholder="+905551234567"${only ? " required" : ""} value="${val}" /></label>`;
    });
    if (extraIds.length && emailLabel) emailLabel.insertAdjacentHTML("afterend", extraIds.join(""));

    if (pwLabel) {
      const noPassword = mode() === "otp";
      pwLabel.hidden = noPassword;
      pwLabel.querySelector("input").required = !noPassword;
      const hint = cfg.password_policy && cfg.password_policy.hint;
      if (hint && !noPassword) pwLabel.insertAdjacentHTML("beforeend", `<small class="passkey-hint">${esc(hint)}</small>`);
    }
    const fields = cfg.registration_schema || [];
    if (fields.length && submit) submit.insertAdjacentHTML("beforebegin", fields.map(fieldHTML).join(""));
    signUpForm.querySelectorAll("select[data-options-url]").forEach((sel) => {
      fetch(sel.dataset.optionsUrl).then((r) => r.json()).then((d) => {
        (d.options || []).forEach((o) => sel.insertAdjacentHTML("beforeend", `<option value="${esc(o.value)}">${esc(o.label)}</option>`));
      }).catch(() => undefined);
    });
    if (prefill) document.getElementById("auth-split-container")?.classList.replace("mode-signin", "mode-signup");
    syncSplitHeight();
  }

  function collectMetadata(fd) {
    const out = {};
    (cfg.registration_schema || []).forEach((f) => {
      const raw = fd.get(`meta_${f.id}`);
      if (f.type === "checkbox") out[f.id] = raw === "on";
      else if (f.type === "number" && raw !== null && raw !== "") out[f.id] = Number(raw);
      else if (raw !== null && raw !== "") out[f.id] = String(raw);
    });
    return out;
  }

  signUpForm?.addEventListener("submit", async (e) => {
    e.preventDefault();
    if (!cfg) return;
    const fd = new FormData(signUpForm);
    const stop = setSubmitLoading(signUpForm, true, t("js.creating", "Creating account…"));
    const body = {
      client_id: clientId,
      email: String(fd.get("email") || "").trim(),
      username: String(fd.get("username") || "").trim(),
      phone: String(fd.get("phone") || "").trim(),
      password: String(fd.get("password") || ""),
      metadata: collectMetadata(fd),
      return_to: new URLSearchParams(location.search).get("return_to") || "",
    };
    try {
      const res = await call("/auth/sign-up", body);
      if (!res.ok) return message(signUpForm, errorFor(res.data), true);
      const identifier = body.email || body.phone || body.username;
      if (mode() === "otp") {
        // Code-only projects: send the first code right away.
        document.getElementById("auth-split-container")?.classList.replace("mode-signup", "mode-signin");
        await sendCode(identifier);
        return;
      }
      const next = new URLSearchParams(location.search);
      next.delete("prefill");
      next.set("message", cfg.email_verify_required && body.email ? t("js.verify_email_message", "Check your email to verify") : t("js.p.account_created", "Account created. Sign in to continue."));
      location.href = "/auth/sign-in?" + next.toString();
    } finally {
      stop();
    }
  });

  fetch("/v1/widget/config?client_id=" + encodeURIComponent(clientId))
    .then((r) => r.json())
    .then((data) => {
      cfg = data;
      renderSignIn();
      renderSignUp();
    })
    .catch(() => {
      window.__sooauthProject = "";
    });
})();
