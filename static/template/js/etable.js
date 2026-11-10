document.addEventListener('DOMContentLoaded', () => {
  const IMYR = window.IMYR

  const MAX_PAGE_VIEW = 6

  IMYR.eTable = (tableDOM, ops) => {
    /**
     * ops:
     * ops.cols = [{ label: string, sortField: string, render: (rowData) => string }]
     * ops.limit = number
     * ops.sortDir = 'desc' || 'asc'
     * ops.fetchData = ({
     *                    limit: number,
     *                    offset: number,
     *                    sortCol: number,
     *                    sortField: string,
     *                    sortDir: 'desc' || 'asc',
     *                  }) => Promise({
     *                    total: number,
     *                    data: [{ 'data' }],
     *                  })
     * ops.onPostRender = () => void
     */

    const configDOMs = {
      thead: tableDOM.getElementsByTagName('thead')[0],
      tbody: tableDOM.getElementsByTagName('tbody')[0],
      caption: tableDOM.getElementsByTagName('caption')[0],
    }

    Object.keys(configDOMs).forEach((key) => {
      if (configDOMs[key]) {
        return
      }

      configDOMs[key] = document.createElement(key)
      tableDOM.appendChild(configDOMs[key])
    })

    const colsConfig = ops.cols || [...configDOMs.thead.getElementsByTagName('th')].map((elem) => {
      return {
        label: elem.innerText,
        render: (data) => data[elem.innerText],
      }
    })

    const sortCols = colsConfig.reduce(
      (arr, col, index) => {
        if (col.sortField) {
          arr.push(index)
        }
        return arr
      },
      [],
    )
  
    const apiParams = {
      limit: ops.limit || 10,
      offset: 0,
    }

    const tData = {
      total: 0,
      data: [],
    }

    if (sortCols.length) {
      apiParams.sortCol = ops.sortCol || sortCols[0]
      apiParams.sortDir = ops.sortDir || 'desc'
      apiParams.sortField = colsConfig[apiParams.sortCol].sortField
    }

    const fetchData = ops.fetchData || ((params) => {
      const queryData = { ...params }
      let url = window.location.pathname

      url += url.includes('?') ? '&' : '?'
      url += Object.entries(queryData).map(([key, value]) => {
        return `${key}=${value}`
      }).join('&')

      return fetch(url).then((res) => res.json()).then((data) => data.data)
    })

    const requestRender = (reset = false) => {
      if (reset) {
        apiParams.offset = 0
      }

      fetchData(apiParams).then((data) => {
        tData.total = data.total
        tData.data = data.data
        render({ colsConfig, ...apiParams, ...tData, ...configDOMs, onPostRender: ops.onPostRender })
      })
    }

    const onClickColHead = (e) => {
      const colIndex = parseInt(e.target.dataset.query.split('_')[1])
      const cConfig = colsConfig[colIndex]

      if (!cConfig.sortField) {
        return
      }

      if (apiParams.sortCol === colIndex) {
        apiParams.sortDir = apiParams.sortDir === 'desc' ? 'asc' : 'desc'
      }

      apiParams.sortCol = colIndex
      apiParams.sortField = cConfig.sortField

      const thDOMS = [...configDOMs.thead.getElementsByTagName('th')]

      thDOMS.forEach((thDOM, index) => {
        thDOM.classList.remove('desc')
        thDOM.classList.remove('asc')

        if (index === colIndex) {
          thDOM.classList.add(apiParams.sortDir)
        }
      })

      requestRender()
    }

    const onClickChangePage = (e) => {
      const page = parseInt(e.target.dataset.info)
      apiParams.offset = (page - 1) * apiParams.limit
      requestRender()
    }

    // INIT

    tableDOM.classList.add('etable')
    configDOMs.caption.innerHTML = 'No data'

    configDOMs.thead.innerHTML = `
    <tr>
      ${colsConfig.map(({ label, sortField }, index) => {
        const style = []
        if (sortField) {
          style.push('sorting')
          if (apiParams.sortCol === index) {
            style.push(apiParams.sortDir)
          }
        }
        return `
        <th class="${style.join(' ')}"
          data-query="th_${index}"
          data-action="sort"
        >
          ${label}
        </th>
        `
      }).join('')}
    </tr>
    `

    // EVENT
  
    tableDOM.addEventListener('click', (e) => {
      if (e.target.dataset.action === 'sort') {
        onClickColHead(e)
      }

      if (e.target.dataset.action === 'change_page') {
        onClickChangePage(e)
      }
    })

    requestRender()

    return {
      configDOMs,
      sortCols,
      colsConfig,
      apiParams,
      tData,
      fetchData,
      requestRender,
    }
  }

  /**
   * render
   */

  const renderBtnPage = (content, active, info) => {
    const renderContent = content || '...'
    const disabled = !active || renderContent === '...'
    return `
    <button type="button"
      class="btn ${renderContent === '...'
                    ? 'btn-link'
                    : disabled
                    ? 'btn-primary'
                    : 'btn-outline-primary'
                  }"
      ${info ? `data-action="change_page" data-info="${info}"` : ''}
      ${disabled ? 'disabled' : ''}
    >
      ${renderContent}
    </button>
    `
  }

  const renderPagesPart = ({ offset, limit, total }) => {
    const currentPage = offset / limit + 1
    const lastPage = Math.ceil(total / limit)

    if (currentPage === 1 || currentPage === lastPage) {
      return `
      ${renderBtnPage(1, currentPage !== 1, 1)}
      ${currentPage === 1 ? renderBtnPage(2, true, 2) : ''}
      ${renderBtnPage('...')}
      ${currentPage === lastPage ? renderBtnPage(lastPage - 1, true, lastPage - 1) : ''}
      ${renderBtnPage(lastPage, currentPage !== lastPage, lastPage)}
      `
    }

    let content = `
      ${renderBtnPage(currentPage - 1, true, currentPage - 1)}
      ${renderBtnPage(currentPage, false, currentPage)}
      ${renderBtnPage(currentPage + 1, true, currentPage + 1)}
    `

    if (currentPage > 3) {
      content = renderBtnPage('...') + content
    }

    if (currentPage < lastPage - 2) {
      content += renderBtnPage('...')
    }

    if (currentPage > 2) {
      content = renderBtnPage(1, true, 1) + content
    }

    if (currentPage < lastPage - 1) {
      content += renderBtnPage(lastPage, true, lastPage)
    }

    return content
  }

  const renderPagesAll = ({ offset, limit, total }) => {
    const currentPage = offset / limit + 1
    const lastPage = Math.ceil(total / limit)

    let content = ''

    for (let i = 0; i < lastPage; i++) {
      const page = i + 1
      content += renderBtnPage(page, page !== currentPage, page)
    }

    return content
  }

  const render = ({ tbody, caption, colsConfig, offset, limit, total, data, onPostRender }) => {
    const currentPage = offset / limit + 1

    const observer = new MutationObserver((mutationList, obs) => {
      onPostRender && onPostRender()
      obs.disconnect()
    })

    observer.observe(tbody, { childList: true })

    caption.innerHTML = `
    <div class="d-flex justify-content-between">
      <div>
        Showing ${data.length === 0 ? 0 : offset + 1} to ${offset + data.length} of totals ${total} entries
      </div>
      <div class="btn-group" role="group">
        <button type="button"
          class="btn btn-outline-primary"
          data-action="change_page"
          data-info="${currentPage - 1}"
          ${offset === 0 && 'disabled'}
        >
          Previous
        </button>
        ${total > MAX_PAGE_VIEW * limit
          ? renderPagesPart({ offset, limit, total })
          : renderPagesAll({ offset, limit, total })
        }
        <button type="button"
          class="btn btn-outline-primary"
          data-action="change_page"
          data-info="${currentPage + 1}"
          ${offset >= total - limit && 'disabled'}
        >
          Next
        </button>
      </div>
    </div>
    `

    tbody.innerHTML = data.map((row) => {
      return `
      <tr>
        ${colsConfig.map(({ render }) => {
          return `
          <td>
            ${render(row)}
          </td>
          `
        }).join('')}
      </tr>
      `
    }).join('')
  }
})
