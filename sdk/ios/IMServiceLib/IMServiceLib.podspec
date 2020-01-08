#
#  Be sure to run `pod spec lint IMServiceLib.podspec' to ensure this is a
#  valid spec and to remove all comments including this before submitting the spec.
#
#  To learn more about Podspec attributes see http://docs.cocoapods.org/specification.html
#  To see working Podspecs in the CocoaPods repo see https://github.com/CocoaPods/Specs/
#

Pod::Spec.new do |s|

  s.name         = "IMServiceLib"
  s.version      = "1.0.0"
  s.summary      = "A message library"
  s.description  = <<-DESC
			        The keeping connection message library.
                   DESC
  s.homepage     = "http://www.huajiao.com"
  s.license      = "MIT"
  s.author       = { "PeterLu" => "lupengyan@hotmail.com" }
  s.platform     = :ios, "8.0"

  s.source       = { :git => "https://git.corp.qihoo.net/maozhua/maozhua_ios.git" }

  s.source_files  = "IMServiceLib/**/*.{h,m,mm,cpp,cc}"

  s.xcconfig = { "HEADER_SEARCH_PATHS" => "${PODS_ROOT}/IMServiceLib/IMServiceLib/IMServiceLib/proto/protobuf-2.6.0/src/google/protobuf/stubs ${PODS_ROOT}/IMServiceLib/IMServiceLib/IMServiceLib/proto/protobuf-2.6.0/src/google/protobuf ${PODS_ROOT}/IMServiceLib/IMServiceLib/IMServiceLib/proto/protobuf-2.6.0/src/google/protobuf/io" }

  s.requires_arc = true

  s.framework    = 'CFNetwork', 'SystemConfiguration', 'UIKit', 'CoreTelephony'

end
