document.addEventListener('DOMContentLoaded', () => {
  const IMYR = window.IMYR
  const A_DAY_MILLISEC = 24 * 60 * 60 * 1000

  IMYR.sheet = {}

  IMYR.sheet.export = (name, data) => {
    const worksheet = XLSX.utils.aoa_to_sheet(data);
    const workbook = XLSX.utils.book_new();
    XLSX.utils.book_append_sheet(workbook, worksheet, name);
    XLSX.writeFileXLSX(workbook, `${name}.xlsx`);
  }

  IMYR.sheet.exportMultiSheets = (name, data) => {
    const workbook = XLSX.utils.book_new();

    Object.entries(data).forEach(([sheetName, sheetData]) => {
      const worksheet = XLSX.utils.aoa_to_sheet(sheetData);
      XLSX.utils.book_append_sheet(workbook, worksheet, sheetName);
    })

    XLSX.writeFileXLSX(workbook, `${name}.xlsx`);
  }

  const sheetToData = (worksheet) => {
    if (!worksheet['!ref']) {
      return []
    }

    const [_from, to] = worksheet['!ref'].split(':')
    const { c: colCount, r: rowCount } = XLSX.utils.decode_cell(to || _from)
    const cols = []
    const rows = []
    const dataArr = []

    for (let i = 0; i <= colCount; i++) {
      cols.push(XLSX.utils.encode_col(i))
    }

    for (let i = 0; i <= rowCount; i++) {
      rows.push(XLSX.utils.encode_row(i))
    }

    rows.forEach((row) => {
      const rowData = []
      dataArr.push(rowData)

      cols.forEach((col) => {
        rowData.push(
          worksheet[`${col}${row}`] ? worksheet[`${col}${row}`].v : '',
        )
      })
    })

    return dataArr
  }

  IMYR.sheet.import = (file, { raw = false } = {}) => {
    return file.arrayBuffer().then((data) => {
      const workbook = XLSX.read(data, { raw })
      const worksheet = workbook.Sheets[workbook.SheetNames[0]]
      return sheetToData(worksheet)
    })
  }

  IMYR.sheet.importSheets = (file, { raw = false } = {}) => {
    return file.arrayBuffer().then((data) => {
      const workbook = XLSX.read(data, { raw })
      const result = {}

      workbook.SheetNames.forEach((sheetName) => {
        const worksheet = workbook.Sheets[sheetName]
        result[sheetName] = sheetToData(worksheet)
      })

      return result
    })
  }

  IMYR.sheet.convertTime = (time) => {
    const day = Math.floor(time) - 25569
    return new Date(day * A_DAY_MILLISEC)
  }
})
