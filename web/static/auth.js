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

var pw = document.getElementById("password");
var confirm = document.getElementById("confirm_password");

if (pw && confirm) {
	function check() {
		confirm.setCustomValidity(
			confirm.value && confirm.value !== pw.value
				? "Passwords do not match"
				: "",
		);
	}
	pw.addEventListener("input", check);
	confirm.addEventListener("input", check);
}
