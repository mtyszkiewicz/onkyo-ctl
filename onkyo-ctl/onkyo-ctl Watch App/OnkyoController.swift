//
//  OnkyoController.swift
//  onkyo-ctl Watch App
//
//  Created by Marcin Tyszkiewicz on 14/04/2024.
//

import Foundation

struct DeviceInfo: Decodable {
    var profile: String
    var volumeLevel: Int
    var subwooferLevel: Int
    var maxVolume: Int
}

struct OnkyoController {
    var apiBaseUrl: String
    public var deviceInfo: DeviceInfo?
    
    init(apiBaseUrl: String) {
        self.apiBaseUrl = apiBaseUrl
    }
    
    func getDeviceInfo() {
        let url = URL(string: "\(self.apiBaseUrl)/device")!
        
        var request = URLRequest(url: url)
        request.httpMethod = "GET"
        
        let task = URLSession.shared.dataTask(with: request) { data, response, error in
            if let error = error {
                print("Error: \(error)")
                return
            }
            
            guard let httpResponse = response as? HTTPURLResponse else {
                print("Invalid HTTP response")
                return
            }

            if httpResponse.statusCode == 200 {
                // Parse JSON data
                if let data = data {
                    do {
                        let result = try JSONDecoder().decode(DeviceInfo.self, from: data)
                        print(result)
                    } catch {
                        print("Error parsing JSON: \(error)")
                    }
                }
            }
        }

        task.resume()
    }

    func selectProfile(name: String) {
        let url = URL(string: "\(self.apiBaseUrl)/profile?name=\(name)")!
        
        var request = URLRequest(url: url)
        request.httpMethod = "PUT"

        let task = URLSession.shared.dataTask(with: request) { data, response, error in
            if let error = error {
                print("Error: \(error)")
                return
            }
            
            guard let httpResponse = response as? HTTPURLResponse else {
                print("Invalid HTTP response")
                return
            }

            if httpResponse.statusCode == 200 {
                // Parse JSON data
                if let data = data {
                    do {
                        let result = try JSONDecoder().decode(DeviceInfo.self, from: data)
                        print(result)
                    } catch {
                        print("Error parsing JSON: \(error)")
                    }
                }
            }
        }

        task.resume()
    }
}
