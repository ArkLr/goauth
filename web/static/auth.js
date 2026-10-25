// Show/hide password toggle.
document.querySelectorAll(".toggle-password").forEach(function (btn) {
	btn.addEventListener("click", function () {
		var input = document.getElementById(btn.dataset.target);
		if (!input) return;

		var show = input.type === "password";
		input.type = show ? "text" : "password";
		btn.classList.toggle("is-visible", show);
		btn.setAttribute("aria-label", show ? "Hide password" : "Show password");
		btn.setAttribute("aria-pressed", String(show));
		input.focus();
	});
});
