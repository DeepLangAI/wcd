const KeyHistoryList = "history";

/*
history = {
    url: "",
    title: "",
    author: "",
    publishTime: "",
    timestamp: int // timestamp
    worthless: bool
}
 */

function BuildHistory(url, title, author, publishTime, worthless = false){
    return {
        url: url,
        title: title,
        author: author,
        publishTime: publishTime,
        timestamp: Date.now(),
        worthless: worthless,
    }
}
function GetHistoryList(){
    let historyList = localStorage.getItem(KeyHistoryList);
    if (historyList == null){
        return []
    }
    historyList = JSON.parse(historyList);
    // sort by time, reverse
    historyList.sort((a, b) => b.timestamp - a.timestamp);
    // format timestamp
    historyList.forEach(item => {
        item.time = new Date(item.timestamp).toLocaleString();
    });
    return historyList
}

function PushHistoryList(history){
    RemoveHistory(history)
    let historyList = GetHistoryList();
    historyList.push(history);
    localStorage.setItem(KeyHistoryList, JSON.stringify(historyList));
}

function RemoveHistory(history){
    let historyList = GetHistoryList();
    historyList = historyList.filter(item => item.url != history.url);
    localStorage.setItem(KeyHistoryList, JSON.stringify(historyList));
}
function RemoveHistoryByUrl(url){
    let historyList = GetHistoryList();
    historyList = historyList.filter(item => item.url != url);
    localStorage.setItem(KeyHistoryList, JSON.stringify(historyList));
}
