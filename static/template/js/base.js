document.addEventListener('DOMContentLoaded', () => {
  const IMYR = { user: null }
  window.IMYR = IMYR

  const TOKEN_KEY = 'jwt'
  const TOKEN_HEADER_KEY = 'jwt'

  if (window.imyrUser && window.imyrUser.username) {
    IMYR.user = window.imyrUser
    IMYR.user.getTokenKey = () => TOKEN_HEADER_KEY

    IMYR.user.getToken = () => {
      return IMYR.getStorage(TOKEN_KEY)
    }
  }

  if (window.imyrData) {
    IMYR.clientData = imyrData
  }

  console.log('IMYR - User', IMYR.user)

  const data = {}

  IMYR.emailRegex = /^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/
  IMYR.separaterRegex = /\n|\r|\s|,|;/gm
  IMYR.storagePrefix = 'imyr__'

  IMYR.getReqHeaders = () => {
    const headers = { 'Content-Type': 'application/json' }
    if (IMYR.user) {
      headers[IMYR.user.getTokenKey()] = IMYR.user.getToken()
    }
    return headers
  }

  IMYR.isObj = (item) => {
    return (typeof item === "object" && !Array.isArray(item) && item !== null)
  }

  IMYR.emptyArr = (arr) => {
    while (arr.length > 0) {
      arr.pop()
    }
  }

  IMYR.twoNum = (num) => {
    return `${num}`.padStart(2, '0')
  }

  IMYR.formatDate = (date) => {
    return `${date.getUTCFullYear()}`
            + `-${IMYR.twoNum(date.getUTCMonth() + 1)}`
            + `-${IMYR.twoNum(date.getUTCDate())}`
  }

  IMYR.formatTime = (date) => {
    new Date().getUTCMinutes()

    return `${IMYR.twoNum(date.getUTCHours())}`
            + `:${IMYR.twoNum(date.getUTCMinutes())}`
  }

  IMYR.formatDatetime = (date) => {
    return `${IMYR.formatTime(date)} ${IMYR.formatDate(date)}`
  }

  IMYR.convertTime = (timeStr) => {
    const date = new Date(timeStr)
    return `${IMYR.twoNum(date.getHours())}`
            + `:${IMYR.twoNum(date.getMinutes())}`
            + ` ${date.getFullYear()}`
            + `-${IMYR.twoNum(date.getMonth() + 1)}`
            + `-${IMYR.twoNum(date.getDate())}`
  }

  IMYR.download = (data, filename, type) => {
    const file = new Blob([data], {type: type});
    if (window.navigator.msSaveOrOpenBlob) {
      // IE10+
      window.navigator.msSaveOrOpenBlob(file, filename)
      return
    }

    const a = document.createElement('a')
    const url = URL.createObjectURL(file)

    a.href = url
    a.download = filename
    document.body.appendChild(a)
    a.click()

    setTimeout(
      () => {
        document.body.removeChild(a)
        window.URL.revokeObjectURL(url);
      },
      0,
    )
  }

  IMYR.setStorage = (key, data) => {
    window.localStorage.setItem(`${IMYR.storagePrefix}${key}`, data)
  }

  IMYR.getStorage = (key) => {
    return window.localStorage.getItem(`${IMYR.storagePrefix}${key}`)
  }

  IMYR.getUrlQuery = (key) => {
    if (!data.urlQueries) {
      data.urlQueries = {}
      if (window.location.search) {
        data.urlQueries = window.location.search.split('?')[1].split('&').reduce(
          (obj, str) => {
            const [k, v] = str.split('=').map((s) => decodeURIComponent(s))
            if (obj[k]) {
              if (!obj[k].length) {
                obj[k] = [obj[k]]
              }

              obj[k].push(v)
            } else {
              obj[k] = v
            }
            return obj
          },
          {},
        )
      }
    }

    return data.urlQueries[key]
  }

  IMYR.setDisable = (btn, isDisable = true) => {
    btn.disabled = isDisable
  }

  IMYR.setLoading = (btn, isLoading = true) => {
    const spinnerDOM = btn.querySelector(`[data-role=spinner]`)
    if (!isLoading) {
      IMYR.setDisable(btn, false)
      spinnerDOM && spinnerDOM.remove()
      return
    }

    IMYR.setDisable(btn, true)

    if (spinnerDOM) {
      return
    }

    const newSpinnerDOM = document.createElement('span')
    newSpinnerDOM.className = 'spinner-border spinner-border-sm'
    newSpinnerDOM.setAttribute('data-role', 'spinner')
    newSpinnerDOM.setAttribute('aria-hidden', 'true')

    btn.appendChild(newSpinnerDOM)
  }

  IMYR.loading = (isLoading, text) => {
    if (isLoading) {
      Swal.isVisible() ? Swal.update({ text }) : Swal.fire({
        icon: 'info',
        text,
        allowOutsideClick: false,
      })
      Swal.showLoading()
    } else {
      Swal.close()
      Swal.hideLoading()
    }
  }

  IMYR.validateEmail = (email) => {
    return !!email.toLowerCase().match(IMYR.emailRegex)
  }

  IMYR.validateOrderNumber = (str) => {
    return /^[a-zA-Z0-9]+-\d+$/.test(str)
  }

  IMYR.request = (action, data) => {
    const headers = { 'Content-Type': 'application/json' }
    if (IMYR.user) {
      headers[IMYR.user.getTokenKey()] = IMYR.user.getToken()
    }

    return fetch('', {
      method: 'POST',
      headers,
      body: JSON.stringify({ action, ...data })
    }).then(
      (res) => res.json(),
    ).catch((error) => {
      console.log('IMYR REQUEST ERROR:', error)
      return {status: 400, message: error.message || error}
    })
  }

  IMYR.fileRequest = (action, formData) => {
    const headers = {}
    if (IMYR.user) {
      headers[IMYR.user.getTokenKey()] = IMYR.user.getToken()
    }

    formData.append('action', action)

    return fetch(`${window.location.pathname}/upload`, {
      method: 'POST',
      headers,
      body: formData,
    }).catch((error) => {
      console.log('IMYR REQUEST ERROR:', error)
      return { status: 400, message: error.message || error }
    })
  }

  IMYR.debounce = (timeout, handler) => {
    let timer = undefined
    return (...args) => {
      clearTimeout(timer)
      timer = setTimeout(() => { handler(...args) }, timeout)
    }
  }

  IMYR.getQueryString = () => {
    if (!window.location.search) {
      return {}
    }

    return window.location.search.replace('?', '').split('&').reduce(
      (obj, splitedStr) => {
        const [key, value] = splitedStr.split('=').map(decodeURIComponent)

        if (obj[key]) {
          if (!Array.isArray(obj[key])) {
            obj[key] = [obj[key]]
          }

          obj[key].push(value)
        } else {
          obj[key] = value
        }

        return obj
      },
      {},
    )
  }

  IMYR.renderObjTable = (value, name) => {
    const renderObj = (item) => {
      return (
        `<table class="table less-table">
          <tbody>
            ${Object.entries(item).map(([k, v]) => (
              `<tr>
                <td>${k}</td>
                <td>${IMYR.renderObjTable(v, k)}</td>
              </tr>`
            )).join('')}
          </tbody>
        </table>`
      )
    }

    const renderArr = (item) => {
      return item.map((v, index) => IMYR.renderObjTable(v, name)).join('<br>')
    }

    if (IMYR.isObj(value)) {
      return renderObj(value)
    }

    if (Array.isArray(value)) {
      return renderArr(value)
    }

    return value
  }

  IMYR.copyClipboard = (text) => {
    navigator.clipboard.writeText(text)
  }

  /**
   * User avatar
   */
  const userAvatar = {
    userColors: [
      '#b94321', '#ec5e32', '#e77524', '#5933a8', '#835cc2', '#60473f', '#8d7469',
      '#3e6e28', '#71a243', '#195145', '#2f8b7e', '#379aa8', '#295b9d', '#3b8ad1',
      '#676fc0', '#bc2c61', '#e44f80', '#4f5f6a', '#7f939e', '#7c28a4', '#aa4fbd',
    ],
    getUserColor: (str) => {
      let num = 0
      for (let i = 0; i < str.length; i++) {
        num += str.charCodeAt(i)
      }

      return userAvatar.userColors[num % userAvatar.userColors.length]
    },
    init: () => {
      const usernameDOM = document.querySelector('[data-query="imyr_base_username"]')
      const avatarDOM = document.querySelector('[data-query="imyr_base_avatar"]')

      if (avatarDOM && usernameDOM && IMYR.user) {
        const username = IMYR.user.username
        usernameDOM.innerText = username
        avatarDOM.style.backgroundColor = userAvatar.getUserColor(username)
        avatarDOM.innerText = username[0]
      }
    }
  }

  /**
   * Sidebar menu state
   */
  const sidebar = {
    init: () => {
      const path = window.location.pathname
      const linkDOMs = [...document.querySelectorAll('[data-query="imyr_sidebar_link"]')]

      const currentLink = linkDOMs.find((dom) => {
        return path === dom.getAttribute('href')
      })

      if (!currentLink) {
        return
      }

      const currentTree = currentLink.closest('[data-query="imyr_sidebar_tree"]')
      const currentGroup = currentLink.closest('[data-query="imyr_sidebar_group"]')

      if (currentTree) {
        currentLink.classList.add('active')
      }

      currentGroup.classList.add('menu-open')
    }
  }


  /**
   * INIT
   */
  userAvatar.init()
  sidebar.init()
})
