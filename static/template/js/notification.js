document.addEventListener('DOMContentLoaded', () => {
  const IMYR = window.IMYR

  const A_DAY_MILLISECONDS = 24 * 60 * 60 * 1000

  const request = (action, data) => {
    const headers = { 'Content-Type': 'application/json' }
    if (IMYR.user) {
      headers[IMYR.user.getTokenKey()] = IMYR.user.getToken()
    }

    return fetch('/notification', {
      method: 'POST',
      headers,
      body: JSON.stringify({ action, ...data })
    }).then(
      (res) => res.json(),
    ).catch((error) => {
      console.log('IMYR NOTI REQUEST ERROR:', error)
      return { status: 400, message: error.message || error }
    })
  }

  const notiDOMs = {
    count: document.getElementById('notiCount'),
    list: document.getElementById('notiList'),
  }

  const typeIconMaps = {
    info: '<i class="fas fa-info-circle mr-2" style="color: #3ca0b7;"></i>',
    warn: '<i class="fas fa-exclamation-circle mr-2" style="color: #f8c328;"></i>',
    error: '<i class="fas fa-times-circle mr-2" style="color: #d33f49;"></i>',
    success: '<i class="fas fa-check-circle mr-2" style="color: #3ca648;"></i>',
    other: '<i class="fas fa-dot-circle mr-2" style="color: #6c757d"></i>',
  }

  const renderNotis = (notis) => {
    if (!notis.length) {
      notiDOMs.count.innerHTML = ''
      notiDOMs.list.innerHTML = '<span class="dropdown-item dropdown-footer">No new notifications</span>'
      return
    }

    notiDOMs.count.innerHTML = notis.length

    notiDOMs.list.innerHTML = notis.map((noti) => {
      const created = new Date(noti.createdAt)
      const today = new Date()
      const timedelta = today - created

      created.setHours(0, 0, 0, 0)
      today.setHours(0, 0, 0, 0)

      const daydelta = (today - created) / A_DAY_MILLISECONDS

      let timeText = ''

      if (timedelta < A_DAY_MILLISECONDS) {
        if (created.getDate() === today.getDate()) {
          timeText = 'Today'
        } else {
          timeText = 'Yesterday'
        }
      } else if (daydelta < 7) {
        timeText = `${daydelta} days ago`
      } else if (daydelta < 30) {
        timeText = `${Math.floor(daydelta / 7)} weeks ago`
      } else {
        timeText = `${Math.floor(daydelta / 30)} months ago`
      }

      return `
      <a href="${noti.url || '#'}"
        class="dropdown-item d-flex justify-content-between align-items-baseline gap-2"
        data-query="noti_item"
        data-info="${noti.id}"
      >
        ${typeIconMaps[noti.type]}
        <p class="flex-grow-1">
          <b>${noti.title}</b><br>
          ${noti.text}
        </p>
        <span class="float-right text-muted text-sm">${timeText}</span>
      </a>
      `
    }).join('<div class="dropdown-divider"></div>')
  }

  notiDOMs.list.addEventListener('click', (e) => {
    const notiDOM = e.target.closest(`[data-query="noti_item"]`)

    if (!notiDOM) {
      return
    }

    const notiId = parseInt(notiDOM.dataset.info)
    request('update_status', {notis: [notiId], status: 'checked'})
  })

  renderNotis([])

  request('get_notifications', {}).then((data) => {
    if (data.status !== 200) {
      return
    }

    renderNotis(data.data)
  })
})
