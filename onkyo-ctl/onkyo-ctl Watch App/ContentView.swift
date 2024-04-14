import SwiftUI



struct ContentView: View {
    var onkyo = OnkyoController(apiBaseUrl: "http://10.205.0.5:8001")
    
    var body: some View {
        VStack(spacing: 10) {
            // First row
            HStack(spacing: 10) {
                Button(action: {
                    onkyo.selectProfile(name: "spotify")
                    print("Music button tapped")
                }) {
                    Text("Music")
                        .padding()
                        .frame(width: 80, height: 65)
                        .background(Color.gray)
                        .foregroundColor(.white)
                        .cornerRadius(10)
                }
                
                Button(action: {
                    onkyo.selectProfile(name: "tv")
                    print("TV button tapped")
                }) {
                    Text("TV")
                        .padding()
                        .frame(width: 80, height: 65)
                        .background(Color.gray)
                        .foregroundColor(.white)
                        .cornerRadius(10)
                }
            }
            
            // Second row
            HStack(spacing: 10) {
                Button(action: {
                    onkyo.selectProfile(name: "vinyl")
                    print("Vinyl button tapped")
                }) {
                    Text("Vinyl")
                        .padding()
                        .frame(width: 80, height: 65)
                        .background(Color.gray)
                        .foregroundColor(.white)
                        .cornerRadius(10)
                }
                
                Button(action: {
                    onkyo.selectProfile(name: "dj")
                    print("DJ button tapped")
                }) {
                    Text("DJ")
                        .padding()
                        .frame(width: 80, height: 65)
                        .background(Color.gray)
                        .foregroundColor(.white)
                        .cornerRadius(10)
                }
            }
        }
        .padding()
    }
}

struct ContentView_Previews: PreviewProvider {
    static var previews: some View {
        ContentView()
    }
}
