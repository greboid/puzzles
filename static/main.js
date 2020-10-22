'use strict';
document.addEventListener("DOMContentLoaded", ready)

let toolResults = null

function ready() {
    document.getElementById('anagramForm').onsubmit = () => {
        handleAnagram()
        return false
    };
    document.getElementById('matchForm').onsubmit = () => {
        handleMatch()
        return false
    };
    document.getElementById('morseForm').onsubmit = () => {
        handleMorse()
        return false
    };
    document.getElementById('t9Form').onsubmit = () => {
        handleT9()
        return false
    };
    document.getElementById('exifUpload').onsubmit = () => {
        handleExifUpload()
        return false
    }
    toolResults = document.getElementById('toolresults')
}

function handleExifUpload() {
    let element = document.getElementById('exifUpload').parentNode
    let photo = document.getElementById("exifFile").files[0]
    let formData = new FormData()
    formData.append("exifFile", photo)
    axios({
        url: '/exifUpload',
        method: "post",
        data: formData,
        headers: {
            'Accept': 'application/json',
            'Content-Type': 'multipart/form-data'
        },
    })
        .then(response => {
            if (!response.data.Success) {
                handleExifResults([], element)
            } else {
                handleExifResults(response.data.Result, element)
            }
        })
        .catch(error => console.log("Error getting exif data: " + error))
}

function handleAnagram() {
    let input = document.getElementById('anagramInput').value
    let element = document.getElementById('anagramForm').parentNode;
    axios.get('/anagram?input='+input)
        .then(function(response){
            if (!response.data.Success) {
                handleResponse([], element)
            } else {
                handleResponse(response.data.Result, element)
            }
        })
        .catch(function (error) {
            console.log("Error getting anagram: " + error)
        })
}

function handleMatch() {
    let input = document.getElementById('matchInput').value
    let element = document.getElementById('matchForm').parentNode;
    axios.get('/match?input='+input)
        .then(function(response){
            if (!response.data.Success) {
                handleResponse([], element)
            } else {
                handleResponse(response.data.Result, element)
            }
        })
        .catch(function (error) {
            console.log("Error getting match: " + error)
        })
}

function handleMorse() {
    let input = document.getElementById('morseInput').value
    let element = document.getElementById('morseForm').parentNode;
    axios.get('/morse?input='+input)
      .then(function(response){
          if (!response.data.Success) {
              handleResponse([], element)
          } else {
              handleResponse(response.data.Result, element)
          }
      })
      .catch(function (error) {
          console.log("Error getting anagram: " + error)
      })
}

function handleT9() {
    let input = document.getElementById('t9Input').value
    let element = document.getElementById('t9Form').parentNode;
    axios.get('/t9?input='+input)
      .then(function(response){
          if (!response.data.Success) {
              handleResponse([], element)
          } else {
              handleResponse(response.data.Result, element)
          }
      })
      .catch(function (error) {
          console.log("Error getting T9: " + error)
      })
}

function handleResponse(results, maxResults = 1000) {
    clearResults(toolResults)
    toolResults.appendChild(createClearResultsButton())
    let resultsList = document.createElement('ul')
    if (results === null) {
    } else if (results.length === 0) {
        resultsList.appendChild(createListItem('No Results'))
    } else if (results.length > maxResults) {
        resultsList.appendChild(createListItem('Over '+maxResults+' results, please narrow down'))
    } else {
        results.forEach(function (result) {
            resultsList.appendChild(createListItem(result))
        })
    }
    toolResults.appendChild(resultsList)
}

function handleExifResults(results) {
    clearResults(toolResults)
    toolResults.appendChild(createClearResultsButton())
    let resultsList = document.createElement('ul')
    if (results === null) {
    } else if (results.length === 0) {
        resultsList.appendChild(createListItem('No Results'))
    } else {
        resultsList.appendChild(createListItem('Size: '+results.width+'x'+results.height))
        resultsList.appendChild(createListItem('Type: '+results.type))
        if (results.exifData.mapLink != null) {
            let listItem = document.createElement('li')
            let link = document.createElement('a');
            link.title = 'Maps Link'
            link.href = results.exifData.mapLink
            link.target = '_blank'
            let linkText = document.createTextNode('Maps link')
            link.appendChild(linkText)
            listItem.appendChild(link)
            resultsList.appendChild(listItem)
        }
        if (results.exifData.datetime != null) {
            resultsList.appendChild(createListItem(results.exifData.datetime))
        }
        if (results.exifData.comments != null) {
            resultsList.appendChild(createListItem(results.exifData.comments))
        }
        if (results.exifData.rawValues == null) {
            resultsList.appendChild(createListItem('No Exif'))
        } else {
            for (const [key, value] of Object.entries(results.exifData.rawValues)) {
                resultsList.appendChild(createListItem(key + ": " + value))
            }
        }
    }
    toolResults.appendChild(resultsList)
}

function createClearResultsButton() {
    let clearResultsButton = document.createElement('span');
    clearResultsButton.appendChild(document.createTextNode('❌'));
    clearResultsButton.id = "clearResults"
    clearResultsButton.onclick = clearResultsAndInputs
    return clearResultsButton
}

function createListItem(content) {
    let listItem = document.createElement('li')
    listItem.appendChild(document.createTextNode(content));
    return listItem
}

function clearResults(element) {
    while (element.firstChild) {
        element.removeChild(element.lastChild)
    }
}

function clearResultsAndInputs() {
    document.getElementById('anagramInput').value = ""
    document.getElementById('morseInput').value = ""
    document.getElementById('matchInput').value = ""
    document.getElementById('t9Input').value = ""
    document.getElementById('flagterms').value = ""
    let exifUpload = document.getElementById('exifUpload')
    exifUpload.innerHTML = exifUpload.innerHTML
    clearResults(toolResults)
}
