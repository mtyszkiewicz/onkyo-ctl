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
    @Published var requestTimeout = false
    
    init(apiBaseUrl: String) {
        self.apiBaseUrl = apiBaseUrl
        updateDeviceInfo()
    }
    
    func updateDeviceInfo() {
        var request = URLRequest(url: URL(string: "\(self.apiBaseUrl)/profile")!)
        request.httpMethod = "GET"
        request.timeoutInterval = 1.0
        sendRequest(request: request)
    }
    
    func selectProfile(name: String) {
        var request = URLRequest(url: URL(string: "\(self.apiBaseUrl)/profile?name=\(name)")!)
        request.httpMethod = "PUT"
        request.timeoutInterval = 1.0
        sendRequest(request: request)
    }
    
    func volumeSet(level: Int) {
        var request = URLRequest(url: URL(string: "\(self.apiBaseUrl)/volume?level=\(level)")!)
        request.httpMethod = "PUT"
        sendRequestNoResponse(request: request)
    }
    
    func powerOn() {
        var request = URLRequest(url: URL(string: "\(self.apiBaseUrl)/power/on")!)
        request.httpMethod = "PUT"
        sendRequestNoResponse(request: request)
    }
    
    func powerOff() {
        var request = URLRequest(url: URL(string: "\(self.apiBaseUrl)/power/off")!)
        request.httpMethod = "PUT"
        sendRequestNoResponse(request: request)
    }
    
    
    func subwooferLevelSet(level: Int) {
        var request = URLRequest(url: URL(string: "\(self.apiBaseUrl)/subwoofer?level=\(level)")!)
        request.httpMethod = "PUT"
        sendRequestNoResponse(request: request)
    }
    
    func sendRequestNoResponse(request: URLRequest) {
        URLSession.shared.dataTask(with: request) { _, _, error in
            if let error = error {
                print("Request failed: \(error.localizedDescription)")
            }
        }.resume()
    }
    
    func sendRequest(request: URLRequest) {
        URLSession.shared.dataTask(with: request) { [weak self] data, response, error in
            DispatchQueue.main.async {
                if let error = error as NSError?, error.code == NSURLErrorTimedOut {
                    self?.requestTimeout = true
                    return
                }
                
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
