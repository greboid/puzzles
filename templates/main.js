'use strict';
document.addEventListener("DOMContentLoaded", ready)

function ready() {
    axios.get('/ideas')
        .then(function(response){
            showCategories(response.data)
            showIdeas(document.getElementById("ideas"), response.data)
        })
        .catch(function(error){
            showCategories(null)
            showIdeas(document.getElementById("ideas"), null)
        })
}

function showCategories(ideas) {
    let radios = document.forms.categories;
    let children = [ ...radios.children ];
    children.forEach(function(child) {
        if (child.tagName !== "FORM") {
            element.removeChild(child)
        }
    })
    let categories = []
    ideas.forEach(function(idea) {
        if (!categories.includes(idea.category)) {
            categories.push(idea.category)
        }
    })
    categories.forEach(function(category) {
        radios.appendChild(htmlToElement('<label><input checked type="checkbox" name="'+category+'" value="'+category+'"/>'+category+'</label>'))
    })
}

function handleCategoryChange(event) {
    console.log(event)
}

function showIdeas(element, ideas) {
    if (ideas === null) {
        element.appendChild(htmlToElement("<p>No Ideas.</p>"))
        return
    }
    let radios = [ ...document.forms.categories].filter(ch => ch.checked ).map(value => value.value)
    let filteredIdeas = ideas.filter(value => radios.includes(value.category))
    let list = htmlToElement("<ul></ul>")
    filteredIdeas.forEach(function(idea) {
        let ideaElement = htmlToElement("<li>"+idea.text+"</li>")
        list.insertAdjacentElement("beforeend", ideaElement)
    })
    element.appendChild(list)
    ideas.forEach(function(idea) {
        if (idea.type === "html+js") {
            addScript(idea.script)
        }
    })
}

function addScript(src) {
    let s = document.createElement( 'script' );
    s.appendChild(document.createTextNode(src))
    document.body.appendChild(s);
}

function handleAnagram() {
    let input = document.forms.anagramForm.elements.anagramInput.value
    let element = document.forms.anagramForm.parentNode;
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
    let element = document.forms.matchForm.parentNode;
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