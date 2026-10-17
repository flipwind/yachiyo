package ymath

import "math"

type Vector3 struct {
	X, Y, Z float64
}

func (vec Vector3) Add(v Vector3) Vector3 {
	return Vector3{
		X: vec.X + v.X,
		Y: vec.Y + v.Y,
		Z: vec.Z + v.Z,
	}
}

func (vec Vector3) Sub(v Vector3) Vector3 {
	return Vector3{
		X: vec.X - v.X,
		Y: vec.Y - v.Y,
		Z: vec.Z - v.Z,
	}
}

func (vec Vector3) Scale(lambda float64) Vector3 {
	return Vector3{
		X: vec.X * lambda,
		Y: vec.Y * lambda,
		Z: vec.Z * lambda,
	}
}

func (vec Vector3) Length() float64 {
	return math.Sqrt(vec.X*vec.X + vec.Y*vec.Y + vec.Z*vec.Z)
}
