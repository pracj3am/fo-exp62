var alertEl = document.querySelectorAll('.alert-danger')[0]
var jiskraEl = document.querySelectorAll('.alert-warning')[0]
var resultEl = document.getElementById('result')
var u0 = document.getElementById('u0')
var i0 = document.getElementById('i0')
var selects = document.querySelectorAll('#aparatura select');
var merak = document.querySelectorAll('#aparatura input[name=merak]');
var colors = {
    '0': ['#ccc', ''],
    '1': ['#000', '#fff'],
    '2': ['#d00', '#fff'],
    '4': ['#fff', '#222'],
    '8': ['#3d3', '#222'],
    '16': ['#06d', '#222'],
};

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
        s = '';
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

            s += '<p>Naměřeno:</p><pre>'+data.Measured.join(',')+'</pre>';
            s += '<p>Napětí na kondenzátoru:</p><pre>'+data.Box.U.join(',')+'</pre>';
            s += '<p>Proud na cívce:</p><pre>'+data.Box.I.join(',')+'</pre>';
            if (data.Box.Jiskra) {
                jiskraEl.style.display = '';
            } else {
                jiskraEl.style.display = 'none';
            }
        }
        resultEl.innerHTML = s;

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
    if (el.value == '0') {
        Array.prototype.forEach.call(el.children, function(opt, i) {
            if (opt.getAttribute('value') != '0') {
                opt.removeAttribute('disabled');
            }
        });
    } else {
        Array.prototype.forEach.call(el.children, function(opt, i) {
            if (opt.getAttribute('value') != '0'
                && opt.getAttribute('value') != el.value)
            { 
                opt.setAttribute('disabled', '');
            }
        });
    }
};

form.addEventListener('submit', function(e) {
    e.preventDefault();
    measure();
});

Array.prototype.forEach.call(selects, function(el, i) {
    el.style.color = colors[el.value][0];
    el.style.backgroundColor = colors[el.value][1];
    endis(el);
    el.addEventListener('change', function(e) {
        e.target.style.color = colors[e.target.value][0];
        e.target.style.backgroundColor = colors[e.target.value][1];
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
