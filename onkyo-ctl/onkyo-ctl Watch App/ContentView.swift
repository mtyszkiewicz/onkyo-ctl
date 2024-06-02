import SwiftUI

struct ProfileButton: View {
    let name: String
    let icon: String
    let isSelected: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            VStack {
                Image(systemName: icon)
                    .font(.title)
                    .foregroundColor(isSelected ? .white : .primary)
                Text(name)
                    .font(.caption)
                    .foregroundColor(isSelected ? .white : .secondary)
            }
            .padding(.horizontal)
            .frame(width: 75, height: 75)
            .background(isSelected ? Color.accentColor : Color.secondary.opacity(0.2))
        }
        .padding(.horizontal)
        .frame(width: 75, height: 75)
        .cornerRadius(12)
        .buttonStyle(PlainButtonStyle())
    }
}


enum Profile: String, CaseIterable {
    case spotify = "spotify"
    case tv = "tv"
    case vinyl = "vinyl"
    case dj = "dj"

    var icon: String {
        switch self {
        case .spotify:
            return "music.note"
        case .tv:
            return "tv"
        case .vinyl:
            return "circle.circle"
        case .dj:
            return "music.mic"
        }
    }
}

struct ContentView: View {
    @StateObject var onkyo = OnkyoController(apiBaseUrl: "http://10.205.0.5:8001")
    @State private var currentVolumeLevel: Double = 10.0
    @State private var previousVolumeLevel: Int = 10

    var body: some View {
        Group {
            if onkyo.requestTimeout {
                VStack(spacing: 20) {
                    Text("There is no need for this app to work outdoors :)")
                        .foregroundColor(.white)
                        .multilineTextAlignment(.center)
                        .padding()

                    Button(action: {
                        reloadDeviceInfo()
                    }) {
                        Text("Reload")
                            .font(.headline)
                            .foregroundColor(.white)
                            .padding()
                            .frame(maxWidth: .infinity)
                            .background(Color.red)
                            .cornerRadius(8)
                    }
                    .padding(.horizontal, 50)
                    .buttonStyle(PlainButtonStyle())
                }
            } else if let deviceInfo = onkyo.deviceInfo {
                deviceContent(deviceInfo)
            } else if onkyo.isDeviceInfoFetched {
                Text("Failed to fetch device information.")
            } else {
                ProgressView("")
            }
        }
        .padding()
    }
    
    private func reloadDeviceInfo() {
        onkyo.requestTimeout = false
        onkyo.isDeviceInfoFetched = false
        onkyo.updateDeviceInfo()
    }
    
    private func deviceContent(_ deviceInfo: DeviceInfo) -> some View {
        VStack(spacing: 10) {
            profileButtons(for: deviceInfo)
            volumeControl(for: deviceInfo)
        }
    }

    private func profileButtons(for deviceInfo: DeviceInfo) -> some View {
        VStack(spacing: 10) {
            ForEach([Array(Profile.allCases.prefix(2)), Array(Profile.allCases.dropFirst(2))], id: \.self) { profiles in
                HStack(spacing: 10) {
                    ForEach(profiles, id: \.self) { profile in
                        ProfileButton(
                            name: profile.rawValue.capitalized,
                            icon: profile.icon,
                            isSelected: deviceInfo.profile == profile.rawValue
                        ) {
                            onkyo.selectProfile(name: profile.rawValue)
                            print("\(profile.rawValue) button tapped")
                        }
                    }
                }
            }
        }
    }

    private func volumeControl(for deviceInfo: DeviceInfo) -> some View {
        Text("Master Volume: \(previousVolumeLevel)")
            .focusable(true)
            .font(.footnote)
            .foregroundColor(Color.secondary.opacity(0.5))
            .digitalCrownRotation(
                $currentVolumeLevel,
                from: 0,
                through: Double(deviceInfo.maxVolume),
                by: 1,
                sensitivity: .low,
                isContinuous: false
            )
            .onChange(of: currentVolumeLevel) { newValue in
                let newVolumeLevel = Int(floor(newValue))
                if newVolumeLevel != previousVolumeLevel {
                    previousVolumeLevel = newVolumeLevel
                    onkyo.volumeSet(level: newVolumeLevel)
                }
            }
            .onChange(of: onkyo.deviceInfo) { newDeviceInfo in
                currentVolumeLevel = Double(newDeviceInfo?.volumeLevel ?? 0)
            }
            .onAppear {
                currentVolumeLevel = Double(deviceInfo.volumeLevel)
                previousVolumeLevel = deviceInfo.volumeLevel
            }
    }
}

struct ContentView_Previews: PreviewProvider {
    static var previews: some View {
        ContentView()
    }
}
