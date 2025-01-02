let selected = null

function dragOver(e) {
    e.preventDefault();
    console.log(e)
    if (isBefore(selected, e.target)) {
        e.target.parentNode.insertBefore(selected, e.target)
    } else {
        e.target.parentNode.insertBefore(selected, e.target.nextSibling)
    }
}

function dragEnd() {
    selected = null
}

function dragEnter(event) {
    // console.log(event)
    event.target.classList.add("isDragging");
    if (event.target.classList.contains("kcol")) {
        event.target.classList.add("dragover");
    }
}

function dragStart(e) {
    e.dataTransfer.effectAllowed = 'move'
    e.dataTransfer.setData('text/plain', null)
    e.target.classList.add("isDragging")
    selected = e.target
}

function isBefore(el1, el2) {
    let cur
    if (el2.parentNode === el1.parentNode) {
        for (cur = el1.previousSibling; cur; cur = cur.previousSibling) {
            if (cur === el2) return true
        }
    }
    return false;
}

function ondrag(ev) {
    ev.preventDefault();
    // console.log(ev)
}

function onDrop(ev) {
    // console.log(ev)
}

function columnOnDragOver(ev) {
    ev.preventDefault()
    if (ev.target.classList.contains("elementx")) { // prevent nesting within element
        return
    }

    // if (ev.target.children.length === 0) {
    //     console.log(ev)
    ev.target.appendChild(selected)
    // }
    ev.target.classList.remove("isOver")
    console.log(ev.target.children.length)
}

function columnOnDragEnter(e) {
    e.target.classList.add("isOver")
}
function columnOnDragLeave(e) {
    e.target.classList.remove("isOver")
}

const taskTemplate = '<div class="elementx" draggable="true" ondragend="dragEnd()" ondrop="onDrop(event)" ondragenter="dragEnter(event)" ondragover="dragOver(event)" ondragstart="dragStart(event)">Oranges 1</div>'

function htmlToNode(html) {
    const template = document.createElement('template');
    template.innerHTML = html
    const nNodes = template.content.childNodes.length
    if (nNodes !== 1) {
        throw new Error(
            `html parameter must represent a single node; got ${nNodes}. ` +
            'Note that leading or trailing spaces around an element in your ' +
            'HTML, like " <img/> ", get parsed as text nodes neighbouring ' +
            'the element; call .trim() on your input to avoid this.'
        )
    }
    return template.content.firstElementChild
}

function taskEl(task) {
    // let taskTpl = '<div class="elementx" draggable="true" ondragend="dragEnd()" ondrop="onDrop(event)" ondragenter="dragEnter(event)" ondragover="dragOver(event)" ondragstart="dragStart(event)">'+
    //     `${task.Title}`+
    //     '</div>'

    let taskTpl = `<div class="task card">
  <div class="card-body elementx">
    <div class="card-header card-title">${task.Title}</div>
<!--    <h6 class="card-subtitle mb-2 text-muted">Card subtitle</h6>-->
    <p class="card-text">${task.DescriptionRendered}</p>
<!--    <a href="#" class="card-link">Card link</a>-->
<!--    <a href="#" class="card-link">Another link</a>-->
  </div>
</div>`


    let el = htmlToNode(taskTpl)
    el.draggable = true
    el.ondragend=dragEnd
    el.ondrop=onDrop
    el.ondragenter=dragEnter
    el.ondragover=dragOver
    el.ondragstart=dragStart
    return  el
}

window.onload = function () {
    // taskTemplate.
    const request = new Request(
        "/column/")

    fetch(request)
        .then((response) => {
            if (response.status === 200) {
                return response.json()
            } else {
                throw new Error("Something went wrong on API server!")
            }
        })
        .then((response) => {
            // console.debug(response)
            let board = document.getElementById("board")

            response.forEach((el) => {
                console.log("el", el)
                let column = document.createElement("div")
                column.ondragenter = columnOnDragEnter
                column.ondragleave = columnOnDragLeave
                column.classList.add("kcol", "col-3", "themed-grid-col")
                column.ondragover = columnOnDragOver
                column.style.background = "#00B0E4"

                el.Tasks.forEach((task) => {
                    // let taskDiv = document.createElement("div")
                    // taskDiv.innerText = task.Title
                    // column.appendChild(taskDiv)
                    column.appendChild(taskEl(task))
                })
                board.appendChild(column)
            })


            // …
        })
        .catch((error) => {
            console.error(error)
        })

}
