htmx.defineExtension('clear-on-event', {
    onEvent : function(name, evt) {
        if (name === "htmx:xhr:loadstart") {
            div = document.getElementById("toolresults")
            while (div.firstChild && div.removeChild(div.firstChild)) {
            }
            div = document.getElementById("flagResults")
            while (div.firstChild && div.removeChild(div.firstChild)) {
            }
        }
    }
})
htmx.defineExtension('disable-on-event', {
    onEvent : function(name, evt) {
        if (name === "htmx:xhr:loadstart") {
            let inputs = document.getElementsByTagName("input");
            for (let i = 0; i < inputs.length; i++) {
                inputs[i].disabled = true;
            }
        }
        if (name === "htmx:xhr:loadend") {
            let inputs = document.getElementsByTagName("input");
            for (let i = 0; i < inputs.length; i++) {
                inputs[i].disabled = false;
            }
        }
    }
})