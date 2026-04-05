pluginManagement {
    repositories {
        google()
        mavenCentral()
        gradlePluginPortal()
    }
}
dependencyResolutionManagement {
    repositoriesMode.set(RepositoriesMode.FAIL_ON_PROJECT_REPOS)
    repositories {
        google()
        mavenCentral()
        // LSPosed/libxposed API repository
        maven { url = uri("https://jitpack.io") }
    }
}



rootProject.name = "Accnotify"
include(":app")
