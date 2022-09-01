function dragOverHandler(ev) {
  ev.preventDefault()
}

function dropHandler(ev, inputId, dropzoneId) {
  ev.preventDefault()

  const inputEl = document.getElementById(inputId)
  if (!inputEl) {
    console.error(`input element not found for ID: ${inputId}`)
    return
  }
  inputEl.files = ev.dataTransfer.files

  const dropzoneEl = document.getElementById(dropzoneId)
  if (!dropzoneEl) {
    console.error(`dropzone element not found for ID: ${dropzoneId}`)
    return
  }
  dropzoneEl.classList.add('border-pink-300')
}
