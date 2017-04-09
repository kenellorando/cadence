<xsl:stylesheet xmlns:xsl="http://www.w3.org/1999/XSL/Transform" version="2.0">
  <xsl:output omit-xml-declaration="yes" method="text" indent="no" media-type="text/javascript" encoding="UTF-8" />
  <xsl:strip-space elements="*" />
  <xsl:template match="/icestats">
    <!-- <xsl:param name="callback" /> <xsl:value-of select="$callback" /> -->parseMusic({<xsl:for-each select="source">"<xsl:value-of select="@mount" />":{"server_name":"<xsl:value-of select="server_name" />","listeners":"<xsl:value-of select="listeners" />","description":"<xsl:value-of select="server_description" />","artist_name":"<xsl:value-of select="artist" />","song_title":"<xsl:value-of select="title" /> ","genre":"<xsl:value-of select="genre" />","bitrate":"<xsl:value-of select="bitrate" />","url":"<xsl:value-of select="server_url" />"}<xsl:if test="position() != last()">
        <xsl:text>,</xsl:text>
      </xsl:if>
    </xsl:for-each>});</xsl:template>
</xsl:stylesheet>
