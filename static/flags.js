// Based on Chris Smith's flagdata demo
// https://github.com/csmith/flagdata

let flags = []
let allWords = []

document.addEventListener('DOMContentLoaded', function () {
    fetch('/flags.json')
        .then(response => response.json())
        .then(json => {
            flags = json
            allWords = flags.reduce(function (previous, current) {
                return previous.concat(current.keywords.filter(k => !previous.includes(k)))
            }, [])
            update()
        })

    const terms = document.getElementById('flagterms')
    const output = document.getElementById('toolresults')
    let lastQuery = []

    function update () {
        const query = terms.value
            .replaceAll('-', ' ')
            .split(' ')
            .map(w => w.toLowerCase().replaceAll(/[^a-z0-9]/g, ''))
            .filter(w => w.length > 0)
            .filter(w => allWords.includes(w))

        if (query === lastQuery) {
            return
        }
        lastQuery = query

        while (output.firstChild) {
            output.removeChild(output.lastChild)
        }

        if (query.length === 0) {
            return
        }

        flags.filter(f => query.every(t => f.keywords.includes(t))).forEach(function (flag) {
            const container = document.createElement('div')
            container.classList.add("flagResult")
            container
                .appendChild(document.createElement('h2'))
                .appendChild(document.createTextNode(flag.country))
            container
                .appendChild(document.createElement('img'))
                .setAttribute('src', "/flags/"+flag.image+".webp")
            output.appendChild(container)
        })
    }

    terms.addEventListener('input', update)
})