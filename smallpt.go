package main

import (
	"math"
	"rand"
	"flag"
	"strconv"
	"fmt"
	"os"
)
var reflCalls uint64

type Vec struct {
	x, y, z float64
}

var NULLVEC *Vec = &Vec{0.0, 0.0, 0.0}

func (this *Vec) String() string {
	return fmt.Sprintf("{%f, %f, %f}", this.x, this.y, this.z)
}

func (this *Vec) Add(v *Vec) *Vec {
	return &Vec{this.x + v.x, this.y + v.y, this.z + v.z}
}

func (this *Vec) Sub(v *Vec) *Vec {
	return &Vec{this.x - v.x, this.y - v.y, this.z - v.z}
}

func (this *Vec) Mul(v *Vec) *Vec {
	return &Vec{this.x * v.x, this.y * v.y, this.z * v.z}
}

func (this *Vec) SMul(f float64) *Vec {
	return &Vec{this.x * f, this.y * f, this.z * f}
}

func (this *Vec) Norm() *Vec {
	return this.SMul(1.0/math.Sqrt(this.x * this.x + this.y * this.y + this.z * this.z))
}

func (this *Vec) Dot(v *Vec) float64 {
	return this.x * v.x + this.y * v.y + this.z * v.z
}

func (this *Vec) Cross(v *Vec) *Vec {
	return &Vec{this.y * v.z - this.z * v.y,
		this.z * v.x - this.x * v.z,
		this.x * v.y - this.y * v.x}
}

type Ray struct {
	Origin, Direction *Vec
}

func (this *Ray) String() string {
	return fmt.Sprintf("{{%s}, {%s}}", this.Origin, this.Direction)
}

const (
	DIFF = iota
	SPEC
	REFR
)

type Sphere struct {
	Radius float64
	Position, Emission, Color *Vec
	Reflection int
}

func (this *Sphere) Intersect(ray *Ray) float64 {
	op := this.Position.Sub(ray.Origin)
	var eps float64 = 1e-4
	var b float64 = op.Dot(ray.Direction)
	var det float64 = b*b - op.Dot(op) + this.Radius * this.Radius
	if det < 0.0 {
		return 0.0
	} else {
		det = math.Sqrt(det)
	}
	if (b - det) > eps {
		return b - det
	} else if (b + det) > eps {
		return b + det
	}
	return 0.0
}

var spheres []*Sphere = []*Sphere{
	&Sphere{1e5, &Vec{ 1e5+1,40.8,81.6}, &Vec{0.0,0.0,0.0},&Vec{.75,.25,.25},  DIFF},//Left
	&Sphere{1e5, &Vec{-1e5+99,40.8,81.6},&Vec{0.0,0.0,0.0},&Vec{.25,.25,.75},  DIFF},//Rght
	&Sphere{1e5, &Vec{50,40.8, 1e5},     &Vec{0.0,0.0,0.0},&Vec{.75,.75,.75},  DIFF},//Back
	&Sphere{1e5, &Vec{50,40.8,-1e5+170}, &Vec{0.0,0.0,0.0},&Vec{0.0, 0.0, 0.0},DIFF},//Frnt
	&Sphere{1e5, &Vec{50, 1e5, 81.6},    &Vec{0.0,0.0,0.0},&Vec{.75,.75,.75},  DIFF},//Botm
	&Sphere{1e5, &Vec{50,-1e5+81.6,81.6},&Vec{0.0,0.0,0.0},&Vec{.75,.75,.75},  DIFF},//Top
	&Sphere{16.5,&Vec{27,16.5,47},       &Vec{0.0,0.0,0.0},&Vec{.999,.999,.999},SPEC},//Mirr
	&Sphere{16.5,&Vec{73,16.5,78},       &Vec{0.0,0.0,0.0},&Vec{.999,.999,.999},REFR},//Glas
	&Sphere{600, &Vec{50,681.6-.27,81.6},&Vec{12,12,12}, &Vec{0.0,0.0,0.0},    DIFF}, //Lite
}

func Clamp(x float64) float64 {
	if x > 1.0 {
		return 1.0
	} else if x < 0.0 {
		return 0.0
	}
	return x
}

func ToByte(x float64) byte {
	return byte(math.Pow(Clamp(x), 1/2.2) * 255 + 0.5)
}

func Intersect(ray *Ray, t *float64, id *int) bool {
	var d, inf float64
	inf = 1e20
	*t = inf
	for i, s := range spheres {
		d = s.Intersect(ray)
		if d != 0.0 && d < *t {
			*t = d
			*id = i
		}
	}
	return *t < inf
}

func Radiance(ray *Ray, depth int) *Vec {
	reflCalls++
	//fmt.Printf("radiance(%s, %d, Xi);\n", ray, depth);
	var t, p float64
	var id int
	if !Intersect(ray, &t, &id) {
		return NULLVEC
	}
	//fmt.Printf("hit = %d\n", id);
	obj := spheres[id]
	x := ray.Origin.Add(ray.Direction.SMul(t))
	n := x.Sub(obj.Position).Norm()
	var nl *Vec
	if n.Dot(ray.Direction) < 0.0 {
		nl = n
	} else {
		nl = n.SMul(-1.0)
	}
	f := obj.Color
	if f.x > f.y && f.x > f.z {
		p = f.x
	} else if f.y > f.z {
		p = f.y
	} else {
		p = f.z
	}
	if depth > 4 {
		if rand.Float64() < p {
			f = f.SMul(1.0/p)
		} else {
			return obj.Emission
		}
	}
	if obj.Reflection == DIFF {
		var r1, r2, r2s float64
		r1 = 2 * math.Pi * rand.Float64()
		r2 = rand.Float64()
		r2s = math.Sqrt(r2)
		var w, u, v, d *Vec
		w = nl
		if math.Fabs(w.x) > 0.1 {
			u = &Vec{0.0, 1.0, 0.0}
		} else {
			u = &Vec{1.0, 0.0, 0.0}
		}
		v = w.Cross(u)
		d = u.SMul(math.Cos(r1) * r2s).Add(v.SMul(math.Sin(r1) * r2s)).Add(w.SMul(math.Sqrt(1 - r2))).Norm()
		return obj.Emission.Add(f.Mul(Radiance(&Ray{x, d}, depth+1)))
	} else if obj.Reflection == SPEC {
		return obj.Emission.Add(f.Mul(Radiance(&Ray{x, ray.Direction.Sub(n.SMul(2*n.Dot(ray.Direction)))}, depth+1)))
	}
	var reflRay *Ray = &Ray{x, ray.Direction.Sub(n.SMul(2*n.Dot(ray.Direction)))}
	var into bool = n.Dot(nl) > 0
	var nc, nt float64 = 1, 1.5
	var nnt, ddn, cos2t float64
	if into {
		nnt = nc / nt
	} else {
		nnt = nt / nc
	}
	ddn = ray.Direction.Dot(nl)
	cos2t = 1 - nnt * nnt * (1 - ddn * ddn)
	if cos2t < 0 {
		return obj.Emission.Add(f.Mul(Radiance(reflRay, depth+1)))
	}
	var tdir *Vec = ray.Direction.SMul(nnt)
	if into {
		tdir = tdir.Sub(n.SMul(ddn*nnt+math.Sqrt(cos2t))).Norm()
	} else {
		tdir = tdir.Sub(n.SMul(-(ddn*nnt+math.Sqrt(cos2t)))).Norm()
	}
	var a, b, c, R0, Re, Tr, P, RP, TP float64
	a, b = nt-nc, nt+nc
	R0 = a*a/(b*b)
	if into {
		c = 1 + ddn
	} else {
		c = 1 - tdir.Dot(n)
	}
	Re = R0 + (1 - R0) *c*c*c*c*c
	Tr = 1 - Re
	P = .25 + .5 * Re
	RP = Re/P
	TP = Tr/(1-P)
	if depth > 1 {
		if rand.Float64() < P {
			return obj.Emission.Add(f.Mul(Radiance(reflRay, depth+1).SMul(RP)))
		} else {
			return obj.Emission.Add(f.Mul(Radiance(&Ray{x, tdir}, depth+1).SMul(TP)))
		}
	}
	return obj.Emission.Add(f.Mul(Radiance(reflRay, depth+1).Add(
		Radiance(&Ray{x, tdir}, depth+1)).SMul(Tr)))
}

func main() {
	flag.Parse()
	var w, h, samps int = 1024, 768, 1
	var cam *Ray = &Ray{&Vec{50,52,295.6}, (&Vec{0, -0.042612, -1}).Norm()}
	var cx, cy, r *Vec
	var c []Vec = make([]Vec, h*w)
	if flag.NArg() > 0 {
		samps, _ = strconv.Atoi(flag.Arg(0))
		samps /= 4
	}
	cx = &Vec{float64(w) * 0.5135 / float64(h), 0.0, 0.0}
	cy = cx.Cross(cam.Direction).Norm().SMul(0.5135)
	for y := 0; y < h; y++ {
		fmt.Printf("\rRendering (%d spp) %5.2f %d",samps*4,100.*float(y)/(float(h)-1.0), reflCalls);
		for x := 0; x < w; x++ {
			for sy, i := 0, (h-y-1)*w+x; sy<2; sy++ {
				for sx := 0; sx<2; sx++ {
					r = new(Vec)
					for s := 0; s < samps; s++ {
						var r1, r2 float64 = 2*rand.Float64(), 2*rand.Float64()
						var dx, dy float64
						if r1 < 1 {
							dx = math.Sqrt(r1) - 1
						} else {
							dx = 1 - math.Sqrt(2-r1)
						}
						if r2 < 1 {
							dy = math.Sqrt(r2) -1
						} else {
							dy = 1 - math.Sqrt(2-r2)
						}
						var d *Vec = cx.SMul(((float64(sx)*.5+dx)/2+float64(x))/float64(w)-.5).Add(
							cy.SMul(((float64(sy)+.5+dy)/2+float64(y))/float64(h)-.5)).Add(cam.Direction)
						r = r.Add(Radiance(&Ray{cam.Origin.Add(d.SMul(140.0)),d.Norm()}, 0).SMul(1.0/float64(samps)))
					}
					c[i] = *(&c[i]).Add((&Vec{Clamp(r.x), Clamp(r.y), Clamp(r.z)}).SMul(0.25))
				}
			}
		}
	}

	f, _ := os.Open("image.ppm", os.O_CREAT|os.O_WRONLY|os.O_TRUNC, 0666)
	f.WriteString(fmt.Sprintf("P3\n%d %d\n%d\n", w, h, 255))
	for _, color := range c {
		f.WriteString(fmt.Sprintf("%d %d %d ", ToByte(color.x), ToByte(color.y), ToByte(color.z)))
	}
	f.Close()
	fmt.Printf("\n%d\n", reflCalls)
}

func main_test() {
	flag.Parse()
	var w, h, samps int = 1024, 768, 1
	var cam *Ray = &Ray{&Vec{50,52,295.6}, (&Vec{0, -0.042612, -1}).Norm()}
	fmt.Printf("Camera = %s\n", cam);
	var cx, cy, r *Vec
	var c []Vec = make([]Vec, h*w)
	if flag.NArg() > 0 {
		samps, _ = strconv.Atoi(flag.Arg(0))
		samps /= 4
	}
	cx = &Vec{float64(w) * 0.5135 / float64(h), 0.0, 0.0}
	cy = cx.Cross(cam.Direction).Norm().SMul(0.5135)
	fmt.Printf("cx = %s, cy = %s\n", cx, cy);
	//for y := 0; y < h; y++ {
	y:=0
		fmt.Printf("\rRendering (%d spp) %5.2f %d",samps*4,100.*float(y)/(float(h)-1.0), reflCalls);
	//for x := 0; x < w; x++ {
	x:=0
	//for sy, i := 0, (h-y-1)*w+x; sy<2; sy++ {
	sy :=0; i:=(h-y-1)*w+x;
	//for sx := 0; sx<2; sx++ {
	sx := 0
	r = new(Vec)
	for s := 0; s < samps; s++ {
		var r1, r2 float64 = 2*rand.Float64(), 2*rand.Float64()
		//r1 = 0.0; r2 = 0.001971;
		var dx, dy float64
		if r1 < 1 {
			dx = math.Sqrt(r1) - 1.0
		} else {
			dx = 1.0 - math.Sqrt(2.0 - r1)
		}
		if r2 < 1 {
			dy = math.Sqrt(r2) - 1.0
		} else {
			dy = 1 - math.Sqrt(2.0 - r2)
		}
		fmt.Printf("r1 = %f, r2 = %f, dx = %f, dy = %f\n", r1, r2, dx, dy)
		var d *Vec = cx.SMul(((float64(sx)*.5+dx)/2.0+float64(x))/float64(w)-.5).Add(
			cy.SMul(((float64(sy)+.5+dy)/2+float64(y))/float64(h)-.5)).Add(cam.Direction)
		fmt.Printf("d = %s\n", d);
		fmt.Printf("r = %s\n", r);
		r = r.Add(Radiance(&Ray{cam.Origin.Add(d.Norm().SMul(140.0)),d.Norm()}, 0).SMul(1.0/float64(samps)))
		fmt.Printf("r = %s\n", r);
	}
	c[i] = *(&c[i]).Add((&Vec{Clamp(r.x), Clamp(r.y), Clamp(r.z)}).SMul(0.25))

	f, _ := os.Open("image.ppm", os.O_CREAT|os.O_WRONLY|os.O_TRUNC, 0666)
	f.WriteString(fmt.Sprintf("P3\n%d %d\n%d\n", w, h, 255))
	for _, color := range c {
		f.Write([]byte{ToByte(color.x), ToByte(color.y), ToByte(color.z)})
	}
	f.Close()
	fmt.Printf("\n%d\n", reflCalls)
}

func main_TEST() {
	var v1, v2, v3 *Vec;
	v1 = &Vec{0.0, 1.0, 0.0};
	v2 = &Vec{1.0, 2.0, 3.0};
	v3 = &Vec{2.0, 2.5, 3.3};

	fmt.Printf("v1 = %s\n", v1);
	fmt.Printf("v2 = %s\n", v2);
	fmt.Printf("v3 = %s\n", v3);
	fmt.Printf("v1 + v2 = %s\n", v1.Add(v2));
	fmt.Printf("v2 + v1 = %s\n", v2.Add(v1));
	fmt.Printf("v1 * v2 = %s\n", v1.Mul(v2));
	fmt.Printf("v2 * v1 = %s\n", v2.Mul(v1));
	fmt.Printf("v1 - v2 = %s\n", v1.Sub(v2));
	fmt.Printf("v2 - v1 = %s\n", v2.Sub(v1));
	fmt.Printf("v3 * -1 = %s\n", v2.SMul(-1));
	fmt.Printf("v1.Norm() = %s\n", v1.Norm());
	fmt.Printf("v2.Norm() = %s\n", v2.Norm());
	fmt.Printf("v1 x v2 = %s\n", v1.Cross(v2));
	fmt.Printf("v1 dot v2 = %f\n", v1.Dot(v2));
	fmt.Printf("v2 dot v1 = %f\n", v2.Dot(v1));
	fmt.Printf("v2 dot v3 = %f\n", v2.Dot(v3));

	var t float64
	var id int
	var ray *Ray;

	ray = &Ray{&Vec{5.920079, 13.487362, 168.423991}, &Vec{-0.314857, -0.275090, -0.908400}}
	fmt.Printf("Intersect(%s) -> ", ray)
	if !Intersect(ray, &t, &id) {
		fmt.Printf("NULL");
	} else {
		fmt.Printf("%d", id);
	}
	fmt.Printf(" == 0?\n");

	ray = &Ray{&Vec{1.031428, 9.216148, 154.319632}, &Vec{0.907061, 0.108974, -0.406652}}
	fmt.Printf("Intersect(%s) -> ", ray)
	if !Intersect(ray, &t, &id) {
		fmt.Printf("NULL");
	} else {
		fmt.Printf("%d", id);
	}
	fmt.Printf(" == 1?\n");

	ray = &Ray{&Vec{98.993889, 20.985367, 110.401268}, &Vec{-0.953139, 0.227378, -0.199562}}
	fmt.Printf("Intersect(%s) -> ", ray)
	if !Intersect(ray, &t, &id) {
		fmt.Printf("NULL");
	} else {
		fmt.Printf("%d", id);
	}
	fmt.Printf(" == 0?\n");

	ray = &Ray{&Vec{1.000407, 44.362437, 89.884022}, &Vec{0.716054, 0.382456, -0.583948}}
	fmt.Printf("Intersect(%s) -> ", ray)
	if !Intersect(ray, &t, &id) {
		fmt.Printf("NULL");
	} else {
		fmt.Printf("%d", id);
	}
	fmt.Printf(" == 5?\n");

	ray = &Ray{&Vec{70.692492, 81.586073, 33.049573}, &Vec{-0.112924, -0.738279, 0.664976}}
	fmt.Printf("Intersect(%s) -> ", ray)
	if !Intersect(ray, &t, &id) {
		fmt.Printf("NULL");
	} else {
		fmt.Printf("%d", id);
	}
	fmt.Printf(" == 7?\n");
}
