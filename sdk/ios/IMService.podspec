#
#  Be sure to run `pod spec lint IMService.podspec' to ensure this is a
#  valid spec and to remove all comments including this before submitting the spec.
#
#  To learn more about Podspec attributes see https://docs.cocoapods.org/specification.html
#  To see working Podspecs in the CocoaPods repo see https://github.com/CocoaPods/Specs/
#

Pod::Spec.new do |spec|

  spec.name         = "IMService"
  spec.version      = "1.0.1"
  spec.summary      = "Instant Messaging service library."
  spec.homepage     = "https://git.huajiao.com/capsules/ios/tree/master/IMSDK"
  spec.license      = "MIT"
  spec.author       = { "PeterLu" => "lupengyan@huajiao.tv" }
  spec.platform     = :ios, '8.0'

  spec.source       = { :git => "git@git.huajiao.com:capsules/ios.git" }


  spec.source_files  = "IMSDK/*.{h}"
  spec.vendored_libraries  = 'IMSDK/libIMServiceLib.a'

  spec.frameworks = 'SystemConfiguration', 'UIKit', 'CoreTelephony', 'CFNetwork', 'Foundation'



  # ――― Project Settings ――――――――――――――――――――――――――――――――――――――――――――――――――――――――― #
  #
  #  If your library depends on compiler flags you can set them in the xcconfig hash
  #  where they will only apply to your library. If you depend on other Podspecs
  #  you can include multiple dependencies to ensure it works.

  # spec.requires_arc = true

  # spec.xcconfig = { "HEADER_SEARCH_PATHS" => "$(SDKROOT)/usr/include/libxml2" }
  # spec.preserve_paths = 'libIMServiceLib.a'
  # spec.dependency "JSONKit", "~> 1.4"

end
