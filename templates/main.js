'use strict';
document.addEventListener("DOMContentLoaded", ready)

function ready() {
    document.forms.anagramForm.onsubmit = () => { handleAnagram(); return false }
    document.forms.matchForm.onsubmit = () => { handleMatch(); return false }
    document.querySelector("#anagramForm span").onclick = function() {
        handleResponse(null, document.getElementById("anagramItem"))
    }
    document.querySelector("#matchForm span").onclick = function() {
        handleResponse(null, document.getElementById("matchItem"))
    }
}

function handleAnagram() {
    let input = document.forms.anagramForm.elements.anagramInput.value
    let element = document.getElementById("anagramItem");
    axios.get('/anagram?input='+input)
        .then(function(response){
            if (!response.data.Success) {
                handleResponse([], element)
            } else {
                handleResponse(response.data.Result, element)
            }
        })
        .catch(function (error) {
            console.log("Error getting anagram")
        })
}

function handleMatch() {
    let input = document.forms.matchForm.elements.matchInput.value
    let element = document.getElementById("matchItem");
    axios.get('/match?input='+input)
        .then(function(response){
            if (!response.data.Success) {
                handleResponse([], element)
            } else {
                handleResponse(response.data.Result, element)
            }
        })
        .catch(function (error) {
            console.log("Error getting anagram")
        })
}

function handleResponse(results, element) {
    let children = [ ...element.children ];
    children.forEach(function(child) {
        if (child.tagName !== "FORM") {
            element.removeChild(child)
        }
    })
    let htmlString = "<ul>"
    if (results === null) {
    } else if (results.length === 0) {
        htmlString += "<li>No Results</li>"
    } else if (results.length > 1000) {
        htmlString += "<li>Over 1000 results, please narrow down</li>"
    } else {
        results.forEach(function (result) {
            htmlString += "<li>" + result + "</li>"
        })
    }
    htmlString += "</ul>"
    element.insertAdjacentElement("beforeend", htmlToElement(htmlString))
}

function htmlToElement(html) {
    let template = document.createElement('template');
    html = html.trim();
    template.innerHTML = html;
    return template.content.firstChild;
}

function htmlToElements(html) {
    let template = document.createElement('template');
    template.innerHTML = html;
    return template.content.childNodes;
}
