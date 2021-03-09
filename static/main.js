var alertEl = document.querySelectorAll('.alert-danger')[0]
var jiskraEl = document.querySelectorAll('.alert-warning')[0]
var u0 = document.getElementById('u0')
var i0 = document.getElementById('i0')
var zapojInpts = document.querySelectorAll('#aparatura .spoj input[type=radio]');
var merak = document.querySelectorAll('#aparatura input[name=merak]');
var form = document.getElementById('aparatura');

var display = new SegmentDisplay("display");
display.pattern         = "####.##";
display.displayAngle    = 0;
display.digitHeight     = 20;
display.digitWidth      = 12;
display.digitDistance   = 2.5;
display.segmentWidth    = 2.0;
display.segmentDistance = 0.5;
display.segmentCount    = 7;
display.cornerType      = 0;
display.colorOn         = "#090909";
display.colorOff        = "#d8d8d8";
display.draw();
display.setValue('');

var intervalId = 0

var measure = function() {
    if (intervalId > 0) {
        clearInterval(intervalId);
        intervalId = 0;
    }

    fetch('/measure', {
        method: 'POST',
        body: new FormData(form),
    }).then(response => response.json())
    .then(data => {
        if (data.Measured) {
            var i = 0;
            intervalId = setInterval(function () {
                var inv = '';
                d = Math.round(100*data.Measured[i])
                if (d < 0) {
                    inv = '-';
                    d = -1 * d;
                }
                d = d.toString().padStart(3, '0');
                d = inv+d.slice(0, -2)+'.'+d.slice(-2);
                if (d.length > 7) {
                    d = 'Erro r';
                } else {
                    d = d.padStart(7, ' ');
                }
                display.setValue(d);
                u0.value = 1000*data.Box.U[i];
                i0.value = data.Box.I[i];
                i++;
            }, 400);

            if (data.Box.Jiskra) {
                jiskraEl.style.display = '';
            } else {
                jiskraEl.style.display = 'none';
            }
        }

        if (data.Msg != '') {
            alertEl.innerHTML = data.Msg;
            alertEl.style.display = '';
        } else {
            alertEl.style.display = 'none';
        }

    }).catch((error) => {
        alertEl.innerHTML = 'Nemůžu se spojit se serverem (' + error + ')';
        alertEl.style.display = '';
    });
};

var endis = function(el) {
    if (el.checked) {
        Array.prototype.forEach.call(el.parentNode.querySelectorAll('label'), function(lbl, i) {
            if (el.value == '0') {
                if (lbl.getAttribute('data-value') != '0') {
                    lbl.style.display = '';
                } else {
                    lbl.style.display = 'none';
                }
            }
            if (el.value != '0') {
                if (lbl.getAttribute('data-value') == '0'
                    || lbl.getAttribute('data-value') == el.value)
                {
                    lbl.style.display = '';
                } else {
                    lbl.style.display = 'none';
                }
            }
        });
    }
};

form.addEventListener('submit', function(e) {
    e.preventDefault();
    measure();
});

Array.prototype.forEach.call(zapojInpts, function(el, i) {
    endis(el);
    el.addEventListener('change', function(e) {
        endis(e.target);
        measure();
    })
});

measure();
Array.prototype.forEach.call(merak, function(el, i) {
    el.addEventListener('change', function(e) {
        measure();
    })
});

var resetBtn = document.getElementById('reset');
var startBtn = document.getElementById('start');
var lapsArea = document.getElementById('laps');

var stopwatch = new SegmentDisplay("stopwatch");
stopwatch.pattern         = "##:##.#";
stopwatch.displayAngle    = 0;
stopwatch.digitHeight     = 20;
stopwatch.digitWidth      = 12;
stopwatch.digitDistance   = 2.5;
stopwatch.segmentWidth    = 2.0;
stopwatch.segmentDistance = 0.5;
stopwatch.segmentCount    = 7;
stopwatch.cornerType      = 0;
stopwatch.colorOn         = "#090909";
stopwatch.colorOff        = "#d8d8d8";
stopwatch.draw();
stopwatch.setValue(' 0:00.0');

var watchInt = 0;
var reset = true;
var startTime = new Date();

var getWatchValue = function() {
    var time    = new Date();
    var elapsed = time - startTime;
    var ms = Math.floor(elapsed % 1000 / 100);
    elapsed = Math.floor(elapsed/1000);
    var seconds = elapsed % 60;
    elapsed = (elapsed - seconds) / 60;
    var minutes = elapsed % 100;

    return ((minutes < 10) ? ' ' : '') + minutes
                + ':' + ((seconds < 10) ? '0' : '') + seconds
                + '.' + ms;
};
var tick = function () {
    stopwatch.setValue(getWatchValue());
};

startBtn.addEventListener('click', function (el) {
    if (el.target.textContent == 'Start') {
        if (reset) {
            startTime = new Date();
            reset = false;
        }
        if (watchInt > 0) {
            clearInterval(watchInt);
        }
        watchInt = setInterval(tick, 100);
        el.target.textContent = 'Stop';
        resetBtn.textContent = 'Lap';
    } else {
        el.target.textContent = 'Start';
        resetBtn.textContent = 'Reset';
        if (watchInt > 0) {
            clearInterval(watchInt);
        }
    }
});

resetBtn.addEventListener('click', function (el) {
    if (el.target.textContent == 'Reset') {
        reset = true;
        stopwatch.setValue(' 0:00.0');
    } else {
        lapsArea.value += getWatchValue() + '\n';
    }
});
