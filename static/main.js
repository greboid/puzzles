'use strict';
document.addEventListener("DOMContentLoaded", ready)

function ready() {
    let toolResults = document.getElementById('toolresults')
    let flagResults = document.getElementById('flagResults')
    document.getElementById('anagramForm').onsubmit = () => {
        clearResults(flagResults)
        handleDictionaryResponse('anagram', document.getElementById('anagramInput'), toolResults)
        return false
    };
    document.getElementById('matchForm').onsubmit = () => {
        clearResults(flagResults)
        handleDictionaryResponse('match', document.getElementById('matchInput'), toolResults)
        return false
    };
    document.getElementById('morseForm').onsubmit = () => {
        clearResults(flagResults)
        handleSimpleResponse('morse', document.getElementById('morseInput'), toolResults)
        return false
    };
    document.getElementById('t9Form').onsubmit = () => {
        clearResults(flagResults)
        handleSimpleResponse('t9', document.getElementById('t9Input'), toolResults)
        return false
    };
    document.getElementById('flagterms').onsubmit = () => {
        clearResults(toolResults)
        return false
    };
    document.getElementById('flagterms').oninput = () => {
        clearResults(toolResults)
        return false
    };
    document.getElementById('analyseForm').onsubmit = () => {
        clearResults(flagResults)
        handleSimpleResponse('analyse', document.getElementById('analyseInput'), toolResults)
        return false
    };
    document.getElementById("exifUpload").onchange = () => {
        clearResults(flagResults)
        handleExifUpload(toolResults)
        return false
    }
}

function handleExifUpload(resultsElement) {
    let photo = document.getElementById("exifFile").files[0]
    let formData = new FormData()
    addLoading(resultsElement)
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
                clearResults(resultsElement)
            } else {
                handleExifResults(response.data.Result, resultsElement)
            }
        })
        .catch(error => handleError(error, resultsElement))
}

function handleSimpleResponse(url, inputElement, resultsElement) {
    addLoading(resultsElement)
    axios.get('/'+url+'?input='+inputElement.value)
        .then(response => handleResponse(response.data, resultsElement))
        .catch(error => handleError(error, resultsElement))
}

function handleDictionaryResponse(url, inputElement, resultsElement) {
    addLoading(resultsElement)
    axios.get('/'+url+'?input='+inputElement.value)
        .then(response => outputDictionaryResults(response.data, resultsElement))
        .catch(error => handleError(error, resultsElement))
}

function outputDictionaryResults(result, resultsElement, maxResults = 1000) {
    clearResults(resultsElement)
    if (!result.Success) {
        resultsElement.appendChild(document.createTextNode('There was no result for this.'))
        return
    }
    let results = result.Result
    resultsElement.appendChild(createClearResultsButton(resultsElement))
    let resultsList = document.createElement('ul')
    if (results === null || results.length === 0) {
        resultsList.appendChild(createListItem('No Results'))
    } else if (results.length > maxResults) {
        resultsList.appendChild(createListItem('Over '+maxResults+' results, please narrow down'))
    } else {
        objectToMap(results).forEach(function (name, dictionary) {
            resultsList.appendChild(createListItem(dictionary))
            let dictionaryList = document.createElement('ul')
            objectToMap(name).forEach(function (word) {
                dictionaryList.appendChild(createListItem(word))
            })
            resultsList.appendChild(dictionaryList)
        })
    }
    resultsElement.appendChild(resultsList)
}

function objectToMap(obj) {
    if (obj == null) {
        return new Map()
    }
    const keys = Object.keys(obj)
    const map = new Map()
    for(let i = 0; i < keys.length; i++){
        map.set(keys[i], obj[keys[i]])
    }
    return map;
}

function handleResponse(result, resultsElement, maxResults = 1000) {
    clearResults(resultsElement)
    if (!result.Success) {
        resultsElement.appendChild(document.createTextNode('There was no result for this.'))
        return
    }
    let results = result.Result
    resultsElement.appendChild(createClearResultsButton(resultsElement))
    let resultsList = document.createElement('ul')
    if (results === null || results.length === 0) {
        resultsList.appendChild(createListItem('No Results'))
    } else if (results.length > maxResults) {
        resultsList.appendChild(createListItem('Over '+maxResults+' results, please narrow down'))
    } else {
        results.forEach(function (result) {
            resultsList.appendChild(createListItem(result))
        })
    }
    resultsElement.appendChild(resultsList)
}

function handleExifResults(results, resultsElement) {
    clearResults(resultsElement)
    resultsElement.appendChild(createClearResultsButton(resultsElement))
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
    resultsElement.appendChild(resultsList)
}

function createClearResultsButton(resultsElement) {
    let clearResultsButton = document.createElement('span');
    clearResultsButton.appendChild(document.createTextNode('❌'));
    clearResultsButton.id = "clearResults"
    clearResultsButton.onclick = () => clearResultsAndInputs(resultsElement)
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

function addLoading(element) {
    while (element.firstChild) {
        element.removeChild(element.lastChild)
    }
    element.appendChild(document.createTextNode('Loading response'))
}

function clearResultsAndInputs(resultsElement) {
    document.getElementById('anagramInput').value = ""
    document.getElementById('morseInput').value = ""
    document.getElementById('matchInput').value = ""
    document.getElementById('t9Input').value = ""
    document.getElementById('flagterms').value = ""
    let exifUpload = document.getElementById('exifUpload')
    exifUpload.innerHTML = exifUpload.innerHTML
    clearResults(resultsElement)
}

function handleError(error, resultsElement) {
    clearResults(resultsElement)
    resultsElement.appendChild(document.createTextNode('Error requesting data: '+error.message))
    console.log(error)
}
