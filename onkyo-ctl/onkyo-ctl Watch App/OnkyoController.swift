import Foundation

struct DeviceInfo: Codable, Equatable {
    var profile: String
    var volumeLevel: Int
    var subwooferLevel: Int
    var maxVolume: Int
}

class OnkyoController: ObservableObject {
    var apiBaseUrl: String
    @Published var deviceInfo: DeviceInfo?
    @Published var isDeviceInfoFetched = false

    init(apiBaseUrl: String) {
        self.apiBaseUrl = apiBaseUrl
        updateDeviceInfo()
    }

    func updateDeviceInfo() {
        var request = URLRequest(url: URL(string: "\(self.apiBaseUrl)/device")!)
        request.httpMethod = "GET"
        sendRequest(request: request)
    }

    func selectProfile(name: String) {
        var request = URLRequest(url: URL(string: "\(self.apiBaseUrl)/profile?name=\(name)")!)
        request.httpMethod = "PUT"
        sendRequest(request: request)
    }

    func volumeSet(level: Int) {
        var request = URLRequest(url: URL(string: "\(self.apiBaseUrl)/volume?level=\(level)")!)
        request.httpMethod = "PUT"
        sendRequest(request: request)
    }

    func sendRequest(request: URLRequest) {
        URLSession.shared.dataTask(with: request) { [weak self] data, response, error in
            DispatchQueue.main.async {
                if let error = error {
                    print("Request failed: \(error.localizedDescription)")
                    return
                }

                guard let httpResponse = response as? HTTPURLResponse else {
                    print("Invalid response")
                    return
                }

                if httpResponse.statusCode != 200 {
                    print("Request failed with status code: \(httpResponse.statusCode)")
                    return
                }

                if let data = data, let deviceInfo = try? JSONDecoder().decode(DeviceInfo.self, from: data) {
                    self?.deviceInfo = deviceInfo
                    self?.isDeviceInfoFetched = true
                } else {
                    print("Unable to decode device info")
                }
            }
        }.resume()
    }
}
