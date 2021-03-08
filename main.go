package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strconv"
)

var templates *template.Template

const (
	Cerna   = 1
	Cervena = 2
	Bila    = 4
	Zelena  = 8
	Modra   = 16

	BilaSeZelenou = (Bila | Zelena) << 5
	BilaSModrou   = (Bila | Modra) << 5
	ZelenaSModrou = (Zelena | Modra) << 5
	Vsechny       = (Bila | Zelena | Modra) << 5

	RL = 100
	Rm = 10e6
	Ue = 9.5
	L  = 0.1
	C  = 1e-6
	dt = 0.4
)

type Data struct {
	Measured []float64
	Box      Box
	Msg      string
}

type Zapojeni struct {
	linkedToCerna   int
	linkedToCervena int
	linkedOnBox     int
	inv             bool
}

type Box struct {
	I, U   []float64
	Jiskra bool
}

func parseKabel(input1, input2 string) (int, int) {
	if input1 == "" {
		log.Println("Neplatný input: chybí první zdířka")
		return 0, 0
	}
	if input2 == "" {
		log.Println("Neplatný input: chybí druhá zdířka")
		return 0, 0
	}
	z1, err := strconv.Atoi(input1)
	if err != nil {
		log.Printf("Neplané číslo zdířky: %s", input1)
	}
	z2, err := strconv.Atoi(input2)
	if err != nil {
		log.Printf("Neplané číslo zdířky: %s", input2)
	}
	if z1 > z2 {
		return z2, z1
	}
	return z1, z2
}

// from <= to assumed
func (z *Zapojeni) addLink(from, to int) {
	if from == 0 || to == 0 {
		// jeden konec nezapojen
		return
	}
	if from == Cerna && to == Cervena {
		// illegal, ignore
		return
	}
	if from == Cerna {
		if z.linkedToCerna > 0 {
			// jeden kabel tam už je
			z.linkedOnBox |= z.linkedToCerna | to
		}
		z.linkedToCerna = to
		if z.linkedToCervena > 0 && z.linkedToCervena < z.linkedToCerna {
			z.inv = true
		}
		return
	}
	if from == Cervena {
		if z.linkedToCervena > 0 {
			// jeden kabel tam už je
			z.linkedOnBox |= z.linkedToCervena | to
		}
		z.linkedToCervena = to
		if z.linkedToCervena < z.linkedToCerna {
			z.inv = true
		}
		return
	}
	z.linkedOnBox |= from | to
}

func (z *Zapojeni) propoj(b int) bool {
	return z.linkedOnBox&b == b
}

func (z *Zapojeni) meri(b int) bool {
	return (z.linkedToCerna|z.linkedToCervena)&b == b
}

func fill(vals []float64, val float64) []float64 {
	for i := range vals {
		vals[i] = val
	}
	return vals
}

func inv(vals []float64) {
	for i := range vals {
		vals[i] = -vals[i]
	}
}

func diff(res []float64, a float64, vals []float64) {
	for i := range vals {
		res[i] = a - vals[i]
		res[i] = math.Round(res[i]*100) / 100
	}
}

func exp1(vals []float64, u0 float64, crop float64) {
	var t float64
	var j int
	a := Ue - u0
	b := -1 / (Rm * C)
	for i := range vals {
		vals[i] = a * math.Exp(b*t)
		if vals[i] < crop {
			j = i
			break
		}
		vals[i] = math.Round(vals[i]*100) / 100
		t += dt
	}
	if j > 0 {
		fill(vals[j:], crop)
	}
}

func exp2(vals []float64, u0 float64) {
	var t float64
	a := u0
	b := -1 / (Rm * C)
	c := L / (Rm * Rm * C)
	d := -Rm / L
	for i := range vals {
		vals[i] = a * (math.Exp(b*t) - c*math.Exp(d*t))
		vals[i] = math.Round(vals[i]*100) / 100
		t += dt
	}
}

func randomDU() float64 {
	dU := rand.NormFloat64() + 1
	if dU < 0 {
		dU = 0
	}
	return math.Round(dU*100) / 100
}

func (z *Zapojeni) measureNothing(box *Box, u0, i0 float64) (msg string) {
	fill(box.U, u0)

	if z.propoj(Zelena | Modra) {
		// cívkou teče proud
		fill(box.I, Ue/RL*1000)
	}
	if z.propoj(Bila | Zelena) {
		// kondenzátor se rychle nabije/vybije
		fill(box.U, Ue)
	}
	if (Bila | Modra) == z.linkedOnBox {
		// část energie cívky se přenese do kondenzátoru
		dU := randomDU()
		if i0 > 0 {
			fill(box.U, u0-32+dU)
		} else if u0 > 0 {
			fill(box.U, -0.47*u0+dU)
		} else {
			fill(box.U, u0)
		}
	}

	if z.linkedOnBox&^(Bila|Zelena) == 0 &&
		i0 > 0 && box.I[0] == 0 {
		box.Jiskra = true
	}

	return
}

func (z *Zapojeni) measureA(box *Box, ma []float64, u0, i0 float64) (msg string) {
	fill(box.U, u0)

	if z.propoj(Zelena|Modra) || z.meri(Zelena|Modra) ||
		(z.meri(Bila|Zelena) && z.propoj(Bila|Modra)) ||
		(z.meri(Bila|Modra) && z.propoj(Bila|Zelena)) {
		// cívkou teče proud
		fill(box.I, Ue/RL*1000)

	}
	if z.propoj(Bila|Zelena) || z.meri(Bila|Zelena) ||
		(z.meri(Bila|Modra) && z.propoj(Zelena|Modra)) ||
		(z.meri(Zelena|Modra) && z.propoj(Bila|Modra)) {
		// kondenzátor se rychle nabije/vybije
		fill(box.U, Ue)
	}

	if (Bila | Modra) == (z.linkedToCerna | z.linkedToCervena | z.linkedOnBox) {
		// část energie cívky se přenese do kondenzátoru
		dU := randomDU()
		if i0 > 0 {
			fill(box.U, u0-32-dU)
		} else if u0 > 0 {
			fill(box.U, -0.47*u0-dU)
		} else {
			fill(box.U, u0)
		}
	}

	if (z.linkedToCerna | z.linkedToCervena) == z.linkedOnBox {
		// nenaměříme nic
	} else if z.meri(Bila | Zelena) { // 1,2
		// merak 0
	} else if z.meri(Zelena | Modra) { // 3,4
		if z.inv {
			fill(ma, -Ue/RL*1000)
			ma[0] = -i0
		} else {
			fill(ma, Ue/RL*1000)
			ma[0] = i0
		}
	} else if z.meri(Bila | Modra) { // 5,6
		if z.propoj(Bila | Zelena) {
			if z.inv {
				fill(ma, -Ue/RL*1000)
				ma[0] = -i0
			} else {
				fill(ma, Ue/RL*1000)
				ma[0] = i0
			}
		} else {
			//merak 0
		}
	}

	if z.linkedOnBox&^(Bila|Zelena) == 0 &&
		(z.linkedToCerna|z.linkedToCervena)&^(Bila|Zelena) == 0 &&
		i0 > 0 && box.I[0] == 0 {
		box.Jiskra = true
	}

	return
}

func (z *Zapojeni) measureV(box *Box, v []float64, u0, i0 float64) (msg string) {
	fill(box.U, u0)

	if z.propoj(Zelena | Modra) {
		// cívkou teče proud
		fill(box.I, Ue/RL*1000)
	}
	if z.propoj(Bila | Zelena) {
		// kondenzátor se rychle nabije/vybije
		fill(box.U, Ue)
	}

	if (Bila | Modra) == z.linkedOnBox {
		// část energie cívky se přenese do kondenzátoru
		dU := randomDU()
		if i0 > 0 {
			fill(box.U, u0-32-dU)
		} else if u0 > 0 {
			fill(box.U, -0.47*u0-dU)
		} else {
			fill(box.U, u0)
		}
	}

	u0 = box.U[0]

	if (z.linkedToCerna | z.linkedToCervena) == z.linkedOnBox {
		// nenaměříme nic
	} else if z.meri(Zelena | Modra) { // 7,8
		if z.inv {
			fill(v, -Ue)
		} else {
			fill(v, Ue)
		}
		v[0] = 0
	} else if z.meri(Bila | Zelena) { // 9,10
		if z.propoj(Bila | Modra) {
			if u0 > 0 {
				msg = "Měřím V(Bila-Zelena) a U_0 > 0"
			}
			// kondenzátor se vybíjí přes cívku
			exp1(v, u0, Ue)
		} else {
			exp1(v, u0, 0)
		}

		diff(box.U, Ue, v)
		if !z.inv {
			inv(v)
		}
	} else if z.meri(Bila | Modra) { // 11,12
		if u0 > 0 { // dioda
			if z.propoj(Bila | Zelena) {
				// zdroj udržuje napětí na kondenzátoru
				copy(v, box.U)
			} else if z.propoj(Zelena | Modra) {
				// cívkou teče 95 mA a je na ní napětí Ue
				// u0 je vždy <= Ue, dioda to nepustí, nenaměřím nic
			} else {
				exp2(v, u0)
				copy(box.U, v)
			}

			if !z.inv {
				inv(v)
			}
		}
	}

	return
}

func Measure(w http.ResponseWriter, r *http.Request) {
	var data Data
	var zap Zapojeni

	r.ParseMultipartForm(1024 * 1024)

	fmt.Println(r.PostForm)

	zap.addLink(parseKabel(r.PostForm.Get("kabel11"), r.PostForm.Get("kabel12")))
	zap.addLink(parseKabel(r.PostForm.Get("kabel21"), r.PostForm.Get("kabel22")))
	zap.addLink(parseKabel(r.PostForm.Get("kabel31"), r.PostForm.Get("kabel32")))

	u0, err := strconv.Atoi(r.PostForm.Get("u0"))
	if err != nil {
		log.Printf("Neplatná hodnota U_0: %s", r.PostForm.Get("u0"))
	}
	i0, err := strconv.Atoi(r.PostForm.Get("i0"))
	if err != nil {
		log.Printf("Neplatná hodnota I_0: %s", r.PostForm.Get("i0"))
	}

	// prapare slices for one hour of measurements
	// TODO: změnit na 9000
	data.Measured = make([]float64, 900)
	data.Box.U = make([]float64, 900)
	data.Box.I = make([]float64, 900)

	if zap.linkedToCerna == 0 || zap.linkedToCervena == 0 {
		data.Msg = zap.measureNothing(&data.Box, float64(u0)/1000, float64(i0))
	} else if r.PostForm.Get("merak") == "A" {
		data.Msg = zap.measureA(&data.Box, data.Measured, float64(u0)/1000, float64(i0))
	} else {
		data.Msg = zap.measureV(&data.Box, data.Measured, float64(u0)/1000, float64(i0))
	}

	b, err := json.Marshal(data)
	if err != nil {
		log.Println(err.Error())
	} else {
		w.Write(b)
	}
}

func GetIndex(w http.ResponseWriter, r *http.Request) {
	if err := templates.ExecuteTemplate(w, "index.html", nil); err != nil {
		log.Println(err.Error())
	}
}

func main() {
	log.Println("Server started")
	http.HandleFunc("/", GetIndex)
	http.HandleFunc("/measure", Measure)
	http.Handle("/static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir("./static"))))
	templates = template.Must(template.ParseGlob("templates/*"))
	log.Println(http.ListenAndServe(":8080", nil))
}
