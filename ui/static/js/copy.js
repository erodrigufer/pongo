const copyIcon = `<svg aria-hidden="true" data-testid="geist-icon" fill="none" height="18" shape-rendering="geometricPrecision" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" viewBox="0 0 24 24" width="18" style="color: currentcolor;"><path d="M8 17.929H6c-1.105 0-2-.912-2-2.036V5.036C4 3.91 4.895 3 6 3h8c1.105 0 2 .911 2 2.036v1.866m-6 .17h8c1.105 0 2 .91 2 2.035v10.857C20 21.09 19.105 22 18 22h-8c-1.105 0-2-.911-2-2.036V9.107c0-1.124.895-2.036 2-2.036z"></path></svg>`

function copyCodeBlock(event) {
	const copyButton = event.currentTarget
	const codeBlock = copyButton.parentElement.querySelector("pre.ssh")
	// Remove any white spaces before or after a string.
	const code = codeBlock.innerText.trim()
	// Remove "$ " prompt at start of lines in code.
	const strippedCode = code.replace(/^[\s]?\$\s+/gm, "")
	window.navigator.clipboard.writeText(strippedCode)

	// Change the button text temporarily.
	copyButton.textContent = "Copied!"
	// After 3 seconds, change back to the copy icon.
	setTimeout(() => copyButton.innerHTML = copyIcon, 3000)
}

// Register event listener for copy button.
const copyButtons = document.querySelectorAll("button.copy")
	copyButtons.forEach((btn) => {
		btn.addEventListener("click", copyCodeBlock)
})
